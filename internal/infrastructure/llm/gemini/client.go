package gemini

import (
	"SQLFactory/pkg/failure"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

type IntentInput struct {
	DBID   string            `json:"db_id"`
	Text   string            `json:"text"`
	Dict   map[string]string `json:"dict,omitempty"`
	Schema any               `json:"schema,omitempty"`
}

type SQLInput struct {
	DBID   string     `json:"db_id"`
	Intent IntentJSON `json:"intent"`
	Schema any        `json:"schema,omitempty"`
}

type IntentJSON struct {
	Intent     string            `json:"intent"`
	Metrics    []string          `json:"metrics,omitempty"`
	Dimensions []string          `json:"dimensions,omitempty"`
	Filters    map[string]string `json:"filters,omitempty"`
	TimeRange  string            `json:"time_range,omitempty"`
}

type ChartSpec struct {
	Type   string   `json:"type"` // none|line|pie|histogram
	X      string   `json:"x,omitempty"`
	Y      []string `json:"y,omitempty"`
	Series string   `json:"series,omitempty"`
}

type SQLJSON struct {
	Title            string    `json:"title"`
	SQL              string    `json:"sql"`
	ExplanationSteps []string  `json:"explanation_steps"`
	Chart            ChartSpec `json:"chart"`
}

func HistoryTitle(userText string, out SQLJSON) string {
	if t := strings.TrimSpace(out.Title); t != "" {
		return t
	}
	t := strings.TrimSpace(userText)
	if t == "" {
		return "Query"
	}
	const maxRunes = 80
	r := []rune(t)
	if len(r) <= maxRunes {
		return t
	}
	return strings.TrimSpace(string(r[:maxRunes])) + "…"
}

type Client struct {
	model  string
	client *genai.Client
}

func NewClient(ctx context.Context, model string) (*Client, error) {
	c, err := genai.NewClient(ctx, nil)
	if err != nil {
		return nil, failure.NewInternalError(err)
	}
	if model == "" {
		model = "gemini-3-flash-preview"
	}
	return &Client{model: model, client: c}, nil
}

func (c *Client) GenerateIntent(ctx context.Context, input IntentInput) (IntentJSON, string, error) {
	schemaJSON, _ := json.Marshal(input.Schema)
	dictJSON, _ := json.Marshal(input.Dict)

	prompt := strings.TrimSpace(fmt.Sprintf(`
You are an assistant that converts a user's natural language analytics request into a compact JSON intent.

Rules:
- Output MUST be valid JSON only. No markdown. No extra text.
- Do not include SQL in this step.

Input:
db_id: %s
user_text: %s
dict: %s
schema: %s

Return JSON with keys:
{
  "intent": string,
  "metrics": string[],
  "dimensions": string[],
  "filters": { string: string },
  "time_range": string
}
`, input.DBID, input.Text, string(dictJSON), string(schemaJSON)))

	out := IntentJSON{}
	raw, err := c.generateJSON(ctx, prompt, &out)
	return out, raw, err
}

func (c *Client) GenerateSQL(ctx context.Context, input SQLInput) (SQLJSON, string, error) {
	schemaJSON, _ := json.Marshal(input.Schema)
	intentJSON, _ := json.Marshal(input.Intent)

	prompt := strings.TrimSpace(fmt.Sprintf(`
You generate a single read-only SQL query for analytics.

Hard rules:
- Output MUST be valid JSON only. No markdown. No extra text.
- Generate exactly ONE statement (no semicolons).
- ONLY SELECT (or WITH ... SELECT). No INSERT/UPDATE/DELETE/DDL.
- Must include LIMIT.

Input:
db_id: %s
intent: %s
schema: %s

Return JSON with keys:
{
  "title": string,
  "sql": string,
  "explanation_steps": string[],
  "chart": {
    "type": "none"|"line"|"pie"|"histogram",
  }
}
`, input.DBID, string(intentJSON), string(schemaJSON)))

	out := SQLJSON{}
	raw, err := c.generateJSON(ctx, prompt, &out)
	return out, raw, err
}

func (c *Client) generateJSON(ctx context.Context, prompt string, out any) (string, error) {
	res, err := c.client.Models.GenerateContent(
		ctx,
		c.model,
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		return "", failure.NewInternalError(err)
	}
	raw := res.Text()
	if raw == "" {
		return "", failure.NewInternalError(errors.New("empty response from gemini"))
	}
	if err := json.Unmarshal([]byte(raw), out); err != nil {
		return raw, failure.NewInvalidRequestError(fmt.Errorf("gemini returned non-json: %w", err))
	}
	return raw, nil
}
