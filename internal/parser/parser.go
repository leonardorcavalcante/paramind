package parser

import (
	"net/url"
	"path"
	"strings"

	"paramind/internal/model"
)

var staticExtensions = map[string]struct{}{
	".jpg":   {},
	".jpeg":  {},
	".png":   {},
	".gif":   {},
	".svg":   {},
	".webp":  {},
	".css":   {},
	".woff":  {},
	".woff2": {},
	".ttf":   {},
	".eot":   {},
	".ico":   {},
	".mp4":   {},
	".mp3":   {},
	".avi":   {},
	".mov":   {},
}

type Parsed struct {
	Canonical string
	Params    []model.QueryParam
}

func ParseLine(line string) (Parsed, bool) {
	parsedURL, err := url.Parse(strings.TrimSpace(line))
	if err != nil {
		return Parsed{}, false
	}

	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" {
		return Parsed{}, false
	}

	if parsedURL.Host == "" {
		return Parsed{}, false
	}

	if parsedURL.RawQuery == "" {
		return Parsed{}, false
	}

	if isStaticAsset(parsedURL.Path) {
		return Parsed{}, false
	}

	params := parseOrderedQuery(parsedURL.RawQuery)
	if len(params) == 0 {
		return Parsed{}, false
	}

	parsedURL.Scheme = scheme
	parsedURL.Host = strings.ToLower(parsedURL.Host)
	parsedURL.Fragment = ""

	return Parsed{
		Canonical: parsedURL.String(),
		Params:    params,
	}, true
}

func isStaticAsset(pathValue string) bool {
	extension := strings.ToLower(path.Ext(pathValue))
	_, blocked := staticExtensions[extension]
	return blocked
}

func parseOrderedQuery(rawQuery string) []model.QueryParam {
	parts := strings.Split(rawQuery, "&")
	params := make([]model.QueryParam, 0, len(parts))

	for _, part := range parts {
		if part == "" {
			continue
		}

		key := part
		value := ""

		if idx := strings.Index(part, "="); idx >= 0 {
			key = part[:idx]
			value = part[idx+1:]
		}

		params = append(params, model.QueryParam{
			Name:  safeQueryUnescape(key),
			Value: safeQueryUnescape(value),
		})
	}

	return params
}

func safeQueryUnescape(value string) string {
	decoded, err := url.QueryUnescape(value)
	if err != nil {
		return value
	}

	return decoded
}
