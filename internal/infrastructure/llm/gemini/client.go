package gemini

import (
	"SQLFactory/internal/config"
	"SQLFactory/internal/domain/service/executor"
	"SQLFactory/pkg/failure"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

type IntentInput struct {
	Text   string            `json:"text"`
	Dict   map[string]string `json:"dict,omitempty"`
	Schema any               `json:"schema,omitempty"`
}

type Client struct {
	model  string
	client *genai.Client
}

func NewClient(ctx context.Context, cfg config.GeminiConfig) (*Client, error) {
	c, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: cfg.ApiKey,
	})
	if err != nil {
		return nil, failure.NewInternalError(err)
	}
	return &Client{model: cfg.Model, client: c}, nil
}

type errorResponse struct {
	Error string `json:"error"`
}

func (c *Client) GenerateSQL(ctx context.Context, request string, dict map[string]string, schema any, dbType string) (*executor.LLMResponse, error) {
	schemaJSON, _ := json.Marshal(schema)
	dictJSON, _ := json.Marshal(dict)

	prompt := strings.TrimSpace(fmt.Sprintf(`
You are an assistant that converts a user's natural language analytics request into a single read-only SQL query for analytics.

First, internally refactor/normalize the user request using the dictionary and schema context. Then generate the SQL.

Hard rules:
- Output MUST be valid JSON only. No markdown. No extra text.
- Generate exactly ONE statement (no semicolons).
- ONLY SELECT (or WITH ... SELECT). No INSERT/UPDATE/DELETE/DDL.
- Must include LIMIT.

Input:
database: %s
user_text: %s
dict: %s
schema: %s

Return JSON with keys:
{
  "title": string,
  "sql": string,
  "explanation_steps": []string,
  "chart_type": "none"|"line"|"pie"|"histogram",
}

Or return an error if the user's request is incorrect:
{
  "error": "..."
}

`, dbType, request, string(dictJSON), string(schemaJSON)))

	res, err := c.client.Models.GenerateContent(
		ctx,
		c.model,
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		return nil, failure.NewInternalError(err)
	}

	rawRes := []byte(res.Text())

	llmErr := new(errorResponse)
	if json.Unmarshal(rawRes, llmErr); llmErr.Error != "" {
		return nil, failure.NewLLMError(llmErr.Error)
	}

	out := new(executor.LLMResponse)
	if err := json.Unmarshal(rawRes, out); err != nil {
		return nil, failure.NewInvalidRequestError(fmt.Errorf("gemini returned non-json: %w", err))
	}
	return out, nil
}

var ErrEmptyResponse = errors.New("empty response from gemini")

func (c *Client) generateJSON(ctx context.Context, prompt string, out any) error {
	res, err := c.client.Models.GenerateContent(
		ctx,
		c.model,
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		return failure.NewInternalError(err)
	}
	raw := res.Text()
	if raw == "" {
		return failure.NewInternalError(ErrEmptyResponse)
	}
	if err := json.Unmarshal([]byte(raw), out); err != nil {
		return failure.NewInvalidRequestError(fmt.Errorf("gemini returned non-json: %w", err))
	}
	return nil
}
