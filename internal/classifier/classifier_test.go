package classifier

import (
	"testing"

	"paramind/internal/model"
)

func TestLocalizedAliases(t *testing.T) {
	t.Parallel()

	classifier := New()

	tests := []struct {
		name       string
		param      string
		wantClass  string
		wantLevel  model.Confidence
		shouldFind bool
	}{
		{
			name:       "pt order alias",
			param:      "ordem",
			wantClass:  "sqli",
			wantLevel:  model.ConfidenceHigh,
			shouldFind: true,
		},
		{
			name:       "pt redirect alias",
			param:      "retorno_url",
			wantClass:  "redirect",
			wantLevel:  model.ConfidenceHigh,
			shouldFind: true,
		},
		{
			name:       "pt auth alias with accent",
			param:      "sess\u00e3o",
			wantClass:  "auth",
			wantLevel:  model.ConfidenceMedium,
			shouldFind: true,
		},
		{
			name:       "pt file alias",
			param:      "arquivo",
			wantClass:  "file",
			wantLevel:  model.ConfidenceHigh,
			shouldFind: true,
		},
		{
			name:       "priority still prefers id over xss",
			param:      "mensagem",
			wantClass:  "id",
			wantLevel:  model.ConfidenceHigh,
			shouldFind: true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			finding, ok := classifier.Classify(test.param, "value")
			if ok != test.shouldFind {
				t.Fatalf("Classify(%q) ok = %v, want %v", test.param, ok, test.shouldFind)
			}

			if !ok {
				return
			}

			if finding.Class != test.wantClass {
				t.Fatalf("Classify(%q) class = %q, want %q", test.param, finding.Class, test.wantClass)
			}

			if finding.Confidence != test.wantLevel {
				t.Fatalf("Classify(%q) confidence = %q, want %q", test.param, finding.Confidence, test.wantLevel)
			}
		})
	}
}
