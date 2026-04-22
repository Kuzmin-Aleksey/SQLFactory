package aiquery

import "SQLFactory/internal/domain/service/sqlrunner"

type Request struct {
	DBID       string
	Text       string
	Dict       map[string]string
	Connection sqlrunner.ConnectionRequest
}

type ChartSpec struct {
	Type   string   `json:"type"` // none|line|pie| histogram
	X      string   `json:"x,omitempty"`
	Y      []string `json:"y,omitempty"`
	Series string   `json:"series,omitempty"`
}

type Response struct {
	SQL              string                 `json:"sql"`
	ExplanationSteps []string               `json:"explanation_steps"`
	Chart            ChartSpec              `json:"chart"`
	Data             *sqlrunner.QueryResult `json:"data"`
}

type IntentJSON struct {
	Intent     string            `json:"intent"`
	Metrics    []string          `json:"metrics,omitempty"`
	Dimensions []string          `json:"dimensions,omitempty"`
	Filters    map[string]string `json:"filters,omitempty"`
	TimeRange  string            `json:"time_range,omitempty"`
}

type SQLJSON struct {
	SQL              string    `json:"sql"`
	ExplanationSteps []string  `json:"explanation_steps"`
	Chart            ChartSpec `json:"chart"`
}
