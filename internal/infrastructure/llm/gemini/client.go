package gemini

import (
	"SQLFactory/internal/domain/service/aiquery"
	"SQLFactory/pkg/failure"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

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

func (c *Client) GenerateIntent(ctx context.Context, input aiquery.IntentInput) (aiquery.IntentJSON, string, error) {
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

	out := aiquery.IntentJSON{}
	raw, err := c.generateJSON(ctx, prompt, &out)
	return out, raw, err
}

func (c *Client) GenerateSQL(ctx context.Context, input aiquery.SQLInput) (aiquery.SQLJSON, string, error) {
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
  "sql": string,
  "explanation_steps": string[],
  "chart": {
    "type": "none"|"line"|"pie"|"histogram",
  }
}
`, input.DBID, string(intentJSON), string(schemaJSON)))

	out := aiquery.SQLJSON{}
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
