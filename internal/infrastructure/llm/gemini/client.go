package gemini

import (
	"SQLFactory/pkg/failure"
	"context"
	"encoding/json"
	"errors"

	"google.golang.org/genai"
)

// Client is a thin wrapper around google.golang.org/genai.
// It is intentionally domain-agnostic: you provide a prompt and an output struct.
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

// GenerateJSON sends prompt to Gemini and unmarshals the response into out.
// It returns raw model text for debugging.
func (c *Client) GenerateJSON(ctx context.Context, prompt string, out any) (raw string, err error) {
	res, err := c.client.Models.GenerateContent(
		ctx,
		c.model,
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		return "", failure.NewInternalError(err)
	}
	raw = res.Text()
	if raw == "" {
		return "", failure.NewInternalError(errors.New("empty response from gemini"))
	}
	if err := json.Unmarshal([]byte(raw), out); err != nil {
		return raw, failure.NewInvalidRequestError(err)
	}
	return raw, nil
}

