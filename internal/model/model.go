package model

type Confidence string

const (
	ConfidenceNone   Confidence = "none"
	ConfidenceLow    Confidence = "low"
	ConfidenceMedium Confidence = "medium"
	ConfidenceHigh   Confidence = "high"

	UnclassifiedClass = "unclassified"
)

type QueryParam struct {
	Name  string
	Value string
}

type Finding struct {
	Param      string     `json:"param"`
	Value      string     `json:"value"`
	Class      string     `json:"class"`
	Confidence Confidence `json:"confidence"`
	Hypotheses []string   `json:"hypotheses,omitempty"`
}

type Result struct {
	URL      string    `json:"url"`
	Findings []Finding `json:"findings,omitempty"`
}

type Stats struct {
	Processed      int
	WithParams     int
	Classified     int
	HighConfidence int
	Duplicates     int
}
