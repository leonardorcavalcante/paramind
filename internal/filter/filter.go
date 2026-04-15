package filter

import (
	"fmt"
	"strings"

	"paramind/internal/model"
)

type Options struct {
	IncludeAll    bool
	MinConfidence model.Confidence
	Categories    map[string]struct{}
}

func ParseMinConfidence(value string) (model.Confidence, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "low":
		return model.ConfidenceLow, nil
	case "medium":
		return model.ConfidenceMedium, nil
	case "high":
		return model.ConfidenceHigh, nil
	default:
		return "", fmt.Errorf("expected one of: low, medium, high")
	}
}

func ParseCategories(value string, known map[string]struct{}) (map[string]struct{}, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}

	categories := make(map[string]struct{})
	for _, raw := range strings.Split(value, ",") {
		category := strings.ToLower(strings.TrimSpace(raw))
		if category == "" {
			continue
		}

		if _, ok := known[category]; !ok {
			return nil, fmt.Errorf("unknown category %q", category)
		}

		categories[category] = struct{}{}
	}

	if len(categories) == 0 {
		return nil, nil
	}

	return categories, nil
}

func Apply(findings []model.Finding, options Options) []model.Finding {
	filtered := make([]model.Finding, 0, len(findings))
	for _, finding := range findings {
		if !passesCategory(finding, options) {
			continue
		}

		if !passesConfidence(finding, options) {
			continue
		}

		filtered = append(filtered, finding)
	}
	return filtered
}

func passesCategory(finding model.Finding, options Options) bool {
	if len(options.Categories) == 0 {
		return true
	}

	if finding.Class == model.UnclassifiedClass {
		return false
	}

	_, ok := options.Categories[strings.ToLower(finding.Class)]
	return ok
}

func passesConfidence(finding model.Finding, options Options) bool {
	if finding.Class == model.UnclassifiedClass {
		return options.IncludeAll && len(options.Categories) == 0
	}

	return confidenceRank(finding.Confidence) >= confidenceRank(options.MinConfidence)
}

func confidenceRank(value model.Confidence) int {
	switch value {
	case model.ConfidenceHigh:
		return 3
	case model.ConfidenceMedium:
		return 2
	case model.ConfidenceLow:
		return 1
	default:
		return 0
	}
}
