package gemini

import (
	"SQLFactory/internal/config"
	"SQLFactory/internal/domain/entity"
	"SQLFactory/internal/infrastructure/llm"
	"SQLFactory/internal/util"
	"SQLFactory/pkg/failure"
	"context"
	"encoding/json"
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

func (c *Client) GenerateSQL(ctx context.Context, llmContext llm.Context, request string, dict map[string]string, schema any, dbType string) (*llm.Response, error) {
	schemaJSON, _ := json.Marshal(schema)
	dictJSON, _ := json.Marshal(dict)

	messages := convertContextToContents(llmContext)

	prompt := strings.TrimSpace(fmt.Sprintf(`
You are an assistant that converts a user's natural language analytics request into a single read-only SQL query for analytics.

First, internally refactor/normalize the user request using the dictionary and schema context. Then generate the SQL.

Hard rules:
- Output MUST be valid JSON only. No markdown. No extra text.
- Generate exactly ONE statement (no semicolons).
- ONLY SELECT (or WITH ... SELECT). No INSERT/UPDATE/DELETE/DDL.
- Must include LIMIT.

Make an additional query to the database (so that you understand what data is there and understand what the user wants) and specify 'need_query: true', and in the sql query that needs to be executed. I will send you the result of the request in the next message.
You can only choose not to use an additional query if you are confident in the query.

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
  "need_query": true | false
}

Or return an error if the user's request is incorrect:
{
  "error": "..."
}

`, dbType, request, string(dictJSON), string(schemaJSON)))

	messages = append(messages, &genai.Content{
		Role:  genai.RoleUser,
		Parts: []*genai.Part{{Text: prompt}},
	})
	llmContext.Append(entity.LLMContext{
		Role:    entity.LLMRoleUser,
		Content: prompt,
	})

	res, err := c.client.Models.GenerateContent(
		ctx,
		c.model,
		messages,
		nil,
	)
	if err != nil {
		return nil, failure.NewInternalError(err)
	}

	rawRes := util.TrimJson([]byte(res.Text()))

	messages = append(messages, &genai.Content{
		Role:  genai.RoleModel,
		Parts: []*genai.Part{{Text: string(rawRes)}},
	})

	llmErr := new(errorResponse)
	if json.Unmarshal(rawRes, llmErr); llmErr.Error != "" {
		return nil, failure.NewLLMError(llmErr.Error)
	}

	out := new(llm.Response)
	if err := json.Unmarshal(rawRes, out); err != nil {
		return nil, failure.NewInternalError(fmt.Errorf("gemini returned non-json: %w", err))
	}
	llmContext.Append(entity.LLMContext{
		Role:    entity.LLMRoleModel,
		Content: string(rawRes),
	})
	out.LLMContext = llmContext
	return out, nil
}

func (c *Client) GenerateSQLSecond(ctx context.Context, llmContext llm.Context, data any) (*llm.Response, error) {
	messages := convertContextToContents(llmContext)

	dataJSON, _ := json.Marshal(data)

	prompt := strings.TrimSpace(fmt.Sprintf(`
Data from DB:
%s

Return result JSON:
{
  "title": string,
  "sql": string,
  "explanation_steps": []string,
  "chart_type": "none"|"line"|"pie"|"histogram",
}


`, string(dataJSON)))

	messages = append(messages, &genai.Content{
		Role:  genai.RoleUser,
		Parts: []*genai.Part{{Text: prompt}},
	})

	llmContext.Append(entity.LLMContext{
		Role:    entity.LLMRoleUser,
		Content: prompt,
	})

	res, err := c.client.Models.GenerateContent(
		ctx,
		c.model,
		messages,
		nil,
	)
	if err != nil {
		return nil, failure.NewInternalError(err)
	}

	out := new(llm.Response)
	if err := json.Unmarshal(util.TrimJson([]byte(res.Text())), out); err != nil {
		return nil, failure.NewInternalError(fmt.Errorf("gemini returned non-json: %w", err))
	}

	llmContext.Append(entity.LLMContext{
		Role:    entity.LLMRoleModel,
		Content: res.Text(),
	})
	out.LLMContext = llmContext
	return out, nil
}

func (c *Client) RegenerateSQL(ctx context.Context, llmContext llm.Context, request string) (*llm.Response, error) {
	messages := convertContextToContents(llmContext)

	prompt := strings.TrimSpace(fmt.Sprintf(`
From user: %s

Return JSON with keys:
{
  "title": string,
  "sql": string,
  "explanation_steps": []string,
  "chart_type": "none"|"line"|"pie"|"histogram",
  "need_query": true | false
}
Don't return error

`, request))

	messages = append(messages, &genai.Content{
		Role:  genai.RoleUser,
		Parts: []*genai.Part{{Text: prompt}},
	})
	llmContext.Append(entity.LLMContext{
		Role:    entity.LLMRoleUser,
		Content: prompt,
	})

	res, err := c.client.Models.GenerateContent(
		ctx,
		c.model,
		messages,
		nil,
	)
	if err != nil {
		return nil, failure.NewInternalError(err)
	}

	rawRes := util.TrimJson([]byte(res.Text()))

	messages = append(messages, &genai.Content{
		Role:  genai.RoleModel,
		Parts: []*genai.Part{{Text: string(rawRes)}},
	})

	out := new(llm.Response)
	if err := json.Unmarshal(rawRes, out); err != nil {
		return nil, failure.NewInternalError(fmt.Errorf("gemini returned non-json: %w", err))
	}
	llmContext.Append(entity.LLMContext{
		Role:    entity.LLMRoleModel,
		Content: string(rawRes),
	})
	out.LLMContext = llmContext
	return out, nil
}

var llmRoleToGenai = map[entity.Role]string{
	entity.LLMRoleUser:  genai.RoleUser,
	entity.LLMRoleModel: genai.RoleModel,
}

func convertContextToContents(llmContext llm.Context) []*genai.Content {
	contents := make([]*genai.Content, len(llmContext))
	for i, contextItem := range llmContext {
		contents[i] = &genai.Content{
			Role:  llmRoleToGenai[contextItem.Role],
			Parts: []*genai.Part{{Text: contextItem.Content}},
		}
	}
	return contents
}
