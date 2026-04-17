package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"paramind/internal/model"
)

type Writer struct {
	out        io.Writer
	jsonOutput bool
	silent     bool
	encoder    *json.Encoder
}

type silentResult struct {
	URL string `json:"url"`
}

func New(out io.Writer, jsonOutput, silent bool) *Writer {
	writer := &Writer{
		out:        out,
		jsonOutput: jsonOutput,
		silent:     silent,
	}
	if jsonOutput {
		writer.encoder = json.NewEncoder(out)
	}
	return writer
}

func (w *Writer) WriteResult(result model.Result) error {
	switch {
	case w.jsonOutput && w.silent:
		return w.encoder.Encode(silentResult{URL: result.URL})
	case w.jsonOutput:
		return w.encoder.Encode(result)
	case w.silent:
		_, err := fmt.Fprintln(w.out, result.URL)
		return err
	default:
		return writeHuman(w.out, result)
	}
}

func WriteStats(out io.Writer, stats model.Stats) error {
	_, err := fmt.Fprintf(
		out,
		"Processed: %d URLs\nWith Params: %d\nClassified: %d\nHigh Confidence: %d\nDuplicates: %d\n",
		stats.Processed,
		stats.WithParams,
		stats.Classified,
		stats.HighConfidence,
		stats.Duplicates,
	)
	return err
}

func writeHuman(out io.Writer, result model.Result) error {
	if _, err := fmt.Fprintf(out, "URL: %s\n\n", result.URL); err != nil {
		return err
	}

	for i, finding := range result.Findings {
		tests := "-"
		if len(finding.Hypotheses) > 0 {
			tests = strings.Join(finding.Hypotheses, ", ")
		}

		if _, err := fmt.Fprintf(
			out,
			"  Param: %s\n  Value: %s\n  Class: %s\n  Confidence: %s\n  Test: %s\n",
			finding.Param,
			finding.Value,
			finding.Class,
			finding.Confidence,
			tests,
		); err != nil {
			return err
		}

		if i < len(result.Findings)-1 {
			if _, err := fmt.Fprintln(out); err != nil {
				return err
			}
		}
	}

	_, err := fmt.Fprintln(out)
	return err
}
