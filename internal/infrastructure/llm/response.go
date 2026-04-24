package llm

type Response struct {
	Title            string   `json:"title"`
	SQL              string   `json:"sql"`
	ExplanationSteps []string `json:"explanation_steps"`
	ChartType        string   `json:"chart_type"`
	NeedQuery        bool     `json:"need_query"`
	LLMContext       Context
}
