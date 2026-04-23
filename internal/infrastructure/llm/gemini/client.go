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

func (c *Client) GenerateRequest(ctx context.Context, request string, dict map[string]string, schema any) (string, error) {
	schemaJSON, _ := json.Marshal(schema)
	dictJSON, _ := json.Marshal(dict)

	prompt := strings.TrimSpace(fmt.Sprintf(`
You are an assistant that converts a user's natural language analytics request into a refactored request.

Rules:
- No markdown. No extra text.
- Do not include SQL in this step.

Input:
user_text: %s
dict: %s
schema: %s

Return only refactored user request.
`, request, string(dictJSON), string(schemaJSON)))

	res, err := c.client.Models.GenerateContent(
		ctx,
		c.model,
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		return "", failure.NewInternalError(err)
	}
	return res.Text(), nil
}

func (c *Client) GenerateSQL(ctx context.Context, request string, schema any, dbType string) (*executor.LLMResponse, error) {
	schemaJSON, _ := json.Marshal(schema)

	prompt := strings.TrimSpace(fmt.Sprintf(`
You generate a single read-only SQL query for analytics.

Hard rules:
- Output MUST be valid JSON only. No markdown. No extra text.
- Generate exactly ONE statement (no semicolons).
- ONLY SELECT (or WITH ... SELECT). No INSERT/UPDATE/DELETE/DDL.
- Must include LIMIT.

Input:
database: %s
user_request: %s
schema: %s

Return JSON with keys:
{
  "title": string,
  "sql": string,
  "explanation_steps": []string,
  "chart_type": "none"|"line"|"pie"|"histogram",
}
`, dbType, request, string(schemaJSON)))

	out := new(executor.LLMResponse)
	if err := c.generateJSON(ctx, prompt, out); err != nil {
		return nil, failure.NewInternalError(err)
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
