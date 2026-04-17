package parser

import (
	"net/url"
	"path"
	"sort"
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
	Signature string
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

	parsedURL.Scheme = scheme
	parsedURL.Host = strings.ToLower(parsedURL.Host)

	rawQuery := parsedURL.RawQuery
	spaPath := ""
	fromFragment := false

	if rawQuery == "" {
		path, query, ok := extractSPAFragment(parsedURL.Fragment)
		if !ok {
			return Parsed{}, false
		}
		rawQuery = query
		spaPath = path
		fromFragment = true
	}

	if isStaticAsset(parsedURL.Path) {
		return Parsed{}, false
	}
	if fromFragment && isStaticAsset(spaPath) {
		return Parsed{}, false
	}

	params := parseOrderedQuery(rawQuery)
	if len(params) == 0 {
		return Parsed{}, false
	}

	sigPath := parsedURL.Path
	var canonical string
	if fromFragment {
		sigPath = parsedURL.Path + "#" + spaPath
		canonical = parsedURL.String()
	} else {
		parsedURL.Fragment = ""
		canonical = parsedURL.String()
	}

	return Parsed{
		Canonical: canonical,
		Signature: buildSignature(parsedURL, sigPath, params),
		Params:    params,
	}, true
}

func extractSPAFragment(fragment string) (spaPath, rawQuery string, ok bool) {
	if fragment == "" {
		return "", "", false
	}
	if !strings.HasPrefix(fragment, "/") && !strings.HasPrefix(fragment, "!") {
		return "", "", false
	}
	idx := strings.Index(fragment, "?")
	if idx < 0 {
		return "", "", false
	}
	return fragment[:idx], fragment[idx+1:], true
}

func buildSignature(u *url.URL, pathPart string, params []model.QueryParam) string {
	buckets := make(map[string]string, len(params))
	order := make([]string, 0, len(params))
	for _, p := range params {
		lower := strings.ToLower(p.Name)
		bucket := valueBucket(p.Value)
		if existing, ok := buckets[lower]; ok {
			if existing != bucket && existing != "m" {
				buckets[lower] = "m"
			}
			continue
		}
		buckets[lower] = bucket
		order = append(order, lower)
	}
	sort.Strings(order)

	parts := make([]string, len(order))
	for i, key := range order {
		parts[i] = key + ":" + buckets[key]
	}
	return u.Scheme + "://" + u.Host + pathPart + "?" + strings.Join(parts, "&")
}

func valueBucket(value string) string {
	if value == "" {
		return "e"
	}
	if isNumeric(value) {
		return "n"
	}
	return "s"
}

func isNumeric(value string) bool {
	hasDigit := false
	dotSeen := false
	for i, r := range value {
		switch {
		case r == '-' && i == 0:
			continue
		case r == '.':
			if dotSeen {
				return false
			}
			dotSeen = true
		case r >= '0' && r <= '9':
			hasDigit = true
		default:
			return false
		}
	}
	return hasDigit
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
