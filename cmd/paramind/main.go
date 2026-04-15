package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"paramind/internal/classifier"
	"paramind/internal/dedupe"
	"paramind/internal/filter"
	"paramind/internal/model"
	"paramind/internal/output"
	"paramind/internal/parser"
)

const (
	scannerInitialBuffer = 1024 * 1024
	scannerMaxBuffer     = 16 * 1024 * 1024
)

func main() {
	var (
		jsonOutput      bool
		includeAll      bool
		minConfidence   string
		silent          bool
		categoryFilters string
	)

	flag.BoolVar(&jsonOutput, "json", false, "emit JSON Lines output")
	flag.BoolVar(&includeAll, "all", false, "include unclassified parameters")
	flag.StringVar(&minConfidence, "min-confidence", "low", "minimum confidence to show: low, medium, high")
	flag.BoolVar(&silent, "silent", false, "only print matching URLs")
	flag.StringVar(&categoryFilters, "category", "", "comma-separated class filter")
	flag.Parse()

	minLevel, err := filter.ParseMinConfidence(minConfidence)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid -min-confidence value: %v\n", err)
		os.Exit(2)
	}

	categorySet, err := filter.ParseCategories(categoryFilters, classifier.KnownCategories())
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid -category value: %v\n", err)
		os.Exit(2)
	}

	classifierEngine := classifier.New()
	filters := filter.Options{
		IncludeAll:    includeAll,
		MinConfidence: minLevel,
		Categories:    categorySet,
	}
	writer := output.New(os.Stdout, jsonOutput, silent)
	seen := dedupe.New()
	stats := model.Stats{}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, scannerInitialBuffer), scannerMaxBuffer)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		stats.Processed++

		parsed, ok := parser.ParseLine(line)
		if !ok {
			continue
		}

		if seen.Seen(parsed.Canonical) {
			continue
		}

		stats.WithParams++

		findings := make([]model.Finding, 0, len(parsed.Params))
		for _, param := range parsed.Params {
			if finding, classified := classifierEngine.Classify(param.Name, param.Value); classified {
				findings = append(findings, finding)
				continue
			}

			if includeAll {
				findings = append(findings, model.Finding{
					Param:      param.Name,
					Value:      param.Value,
					Class:      model.UnclassifiedClass,
					Confidence: model.ConfidenceNone,
				})
			}
		}

		filtered := filter.Apply(findings, filters)
		if len(filtered) == 0 {
			continue
		}

		for _, finding := range filtered {
			if finding.Confidence == model.ConfidenceHigh {
				stats.HighConfidence++
			}
		}

		stats.Classified++

		if err := writer.WriteResult(model.Result{
			URL:      parsed.Canonical,
			Findings: filtered,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "write error: %v\n", err)
			os.Exit(1)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "stdin read error: %v\n", err)
		os.Exit(1)
	}

	if err := output.WriteStats(os.Stderr, stats); err != nil {
		fmt.Fprintf(os.Stderr, "stats write error: %v\n", err)
		os.Exit(1)
	}
}
