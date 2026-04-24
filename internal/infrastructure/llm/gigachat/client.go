package gigachat

import (
	"SQLFactory/internal/config"
	"SQLFactory/internal/domain/entity"
	"SQLFactory/internal/infrastructure/llm"
	"SQLFactory/internal/util"
	"SQLFactory/pkg/failure"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	masterToken   string
	accessToken   string
	expiresAt     time.Time
	GigaChatModel string
}

type UpdateTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"`

	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

func NewClient(cfg config.GigaChatConfig) (*Client, error) {
	client := &Client{
		masterToken:   cfg.ApiKey,
		GigaChatModel: cfg.Model,
	}

	if err := client.updateToken(); err != nil {
		return nil, err
	}

	go func() {
		for {
			time.Sleep(client.expiresAt.Sub(time.Now()))
			if err := client.updateToken(); err != nil {
				log.Printf("Error updating token: %s", err.Error())
				continue
			}

		}
	}()

	return client, nil
}

type errorResponse struct {
	Error string `json:"error"`
}

func (c *Client) GenerateSQL(ctx context.Context, llmContext llm.Context, request string, dict map[string]string, schema any, dbType string) (*llm.Response, error) {
	schemaJSON, _ := json.Marshal(schema)
	dictJSON, _ := json.Marshal(dict)

	messages := convertContextToAIMessages(llmContext)

	prompt := strings.TrimSpace(fmt.Sprintf(`
Ты — аналитический ассистент, который преобразует запросы пользователей на естественном языке в SQL и предлагает подходящий тип визуализации. На вход подаются:
- 'user_query' — текст запроса пользователя;
- 'db_schema' — описание схемы базы данных (таблицы, поля, связи);
- 'db_type' — строка с типом СУБД (например, "PostgreSQL", "MySQL", "ClickHouse" и т. п.).
- 'dict' — Сопоставление [Сленг/Синоним] -> [Каноническое название поля/таблицы].

Твоя задача:
1. Понять, какую аналитику хочет получить пользователь.
2. Определить наиболее подходящий тип графика: 
   - "line" — временные ряды, тренды, изменения во времени;
   - "pie" — сравнение долей, распределение по категориям;
   - "histogram" — распределение числовых значений по интервалам;
   - "none" — если запрос не предполагает график (например, получение сырых данных, списка записей, единичного показателя без категоризации).
3. Выполни дополнительный запрос к бд (чтобы ты понимал какие данные там лежат и понимал что хочет пользователь) и укажи 'need_query: true', а в sql запрос который нужно выполнить. Следующим сообщением я верну тебе результат запроса.
Если дополнительный запрос не требуется то выполни следующие шаги.
	4. Сгенерировать один корректный и безопасный SQL-запрос (только SELECT), который вернёт данные именно в том виде, который нужен для построения выбранного графика. Учти диалект SQL, поддерживаемый указанной 'db_type'', используй подходящие функции и синтаксис.
	5. Сформировать понятный заголовок графика.
	6. Описать шаги рассуждения — как ты пришёл к типу графика и SQL.

Так как ты не видешь данные в бд - сделай дополнительный запрос чтобы ты смог более точно составить итоговый отет в следующем запросе.

Требования к SQL:
- Если ты готов сразу составить запрос то он должен быть готов к выполнению и возвращать колонки, необходимые для построения графика (например, для line — колонка со временем/осью X и числовая колонка; для pie — категориальная и числовая; для histogram — интервал или границы и частота).
- Используй агрегации (SUM, COUNT, AVG), группировку, сортировку, фильтрацию по необходимости.
- Избегай DML, DDL, ввода произвольных строк от пользователя; запрос должен быть параметризован только логикой самого вопроса.

Ответ верни исключительно в виде JSON-объекта с полями:
{
  "title": "строка с заголовком графика",
  "sql": "строка с SQL-запросом",
  "explanation_steps": ["шаг 1", "шаг 2", ...],
  "chart_type": "none" | "line" | "pie" | "histogram",
  "need_query": true | false
}

Или верни ошибку если запрос пользователя некорректен:
{
  "error": "..."
}

Не добавляй никаких пояснений вне JSON. Строки внутри JSON экранируй должным образом. Весь ответ должен быть валидным JSON.

Теперь примени эти правила к следующим входным данным:
user_request: %s
dict: %s
db_schema: %s
db_type: %s
`, request, string(dictJSON), string(schemaJSON), dbType))

	messages = append(messages, AIMessage{
		Role:    RoleUser,
		Content: prompt,
	})
	llmContext.Append(entity.LLMContext{
		Role:    entity.LLMRoleUser,
		Content: prompt,
	})

	res, err := c.request(ctx, messages, 0.6)
	if err != nil {
		return nil, failure.NewInternalError(err)
	}

	messages = append(messages, AIMessage{
		Role:    RoleAssistant,
		Content: res,
	})

	rawRes := util.TrimJson([]byte(res))

	llmErr := new(errorResponse)
	if json.Unmarshal(rawRes, llmErr); llmErr.Error != "" {
		return nil, failure.NewLLMError(llmErr.Error)
	}

	out := new(llm.Response)
	if err := json.Unmarshal(rawRes, out); err != nil {
		return nil, failure.NewInternalError(fmt.Errorf("gigachat returned non-json: %w", err))
	}

	llmContext.Append(entity.LLMContext{
		Role:    entity.LLMRoleModel,
		Content: res,
	})
	out.LLMContext = llmContext

	return out, nil
}

func (c *Client) GenerateSQLSecond(ctx context.Context, llmContext llm.Context, data any) (*llm.Response, error) {
	messages := convertContextToAIMessages(llmContext)

	dataJSON, _ := json.Marshal(data)

	prompt := strings.TrimSpace(fmt.Sprintf(`
Данные из БД:
%s

Не добавляй никаких пояснений вне JSON. Строки внутри JSON экранируй должным образом. Весь ответ должен быть валидным JSON

Верни итоговый валидный JSON:
{
  "title": "строка с заголовком графика",
  "sql": "строка с SQL-запросом",
  "explanation_steps": ["шаг 1", "шаг 2", ...],
  "chart_type": "none" | "line" | "pie" | "histogram"
}
`, string(dataJSON)))

	messages = append(messages, AIMessage{
		Role:    RoleUser,
		Content: prompt,
	})
	llmContext.Append(entity.LLMContext{
		Role:    entity.LLMRoleUser,
		Content: prompt,
	})

	res, err := c.request(ctx, messages, 0.4)
	if err != nil {
		return nil, failure.NewInternalError(err)
	}

	out := new(llm.Response)
	if err := json.Unmarshal(util.TrimJson([]byte(res)), out); err != nil {
		return nil, failure.NewInternalError(fmt.Errorf("gigachat returned non-json (%s): %w", res, err))
	}

	llmContext.Append(entity.LLMContext{
		Role:    entity.LLMRoleModel,
		Content: res,
	})
	out.LLMContext = llmContext
	return out, nil
}

var llmRoleToGigachat = map[entity.Role]string{
	entity.LLMRoleUser:  RoleUser,
	entity.LLMRoleModel: RoleAssistant,
}

func convertContextToAIMessages(llmContext llm.Context) []AIMessage {
	messages := make([]AIMessage, len(llmContext))
	for i, contextItem := range llmContext {
		messages[i] = AIMessage{
			Role:    llmRoleToGigachat[contextItem.Role],
			Content: contextItem.Content,
		}
	}
	return messages
}

func (c *Client) RegenerateSQL(ctx context.Context, llmContext llm.Context, request string) (*llm.Response, error) {
	messages := convertContextToAIMessages(llmContext)

	prompt := strings.TrimSpace(fmt.Sprintf(`
Уточнение от пользователья: %s

Ответ верни исключительно в виде JSON-объекта с полями:
{
  "title": "строка с заголовком графика",
  "sql": "строка с SQL-запросом",
  "explanation_steps": ["шаг 1", "шаг 2", ...],
  "chart_type": "none" | "line" | "pie" | "histogram",
  "need_query": true | false
}

Не возвращай ошибку

Не добавляй никаких пояснений вне JSON. Строки внутри JSON экранируй должным образом. Весь ответ должен быть валидным JSON.
`, request))

	messages = append(messages, AIMessage{
		Role:    RoleUser,
		Content: prompt,
	})
	llmContext.Append(entity.LLMContext{
		Role:    entity.LLMRoleUser,
		Content: prompt,
	})

	res, err := c.request(ctx, messages, 0.6)
	if err != nil {
		return nil, failure.NewInternalError(err)
	}

	messages = append(messages, AIMessage{
		Role:    RoleAssistant,
		Content: res,
	})

	rawRes := util.TrimJson([]byte(res))

	llmErr := new(errorResponse)
	if json.Unmarshal(rawRes, llmErr); llmErr.Error != "" {
		return nil, failure.NewLLMError(llmErr.Error)
	}

	out := new(llm.Response)
	if err := json.Unmarshal(rawRes, out); err != nil {
		return nil, failure.NewInternalError(fmt.Errorf("gigachat returned non-json: %w", err))
	}

	llmContext.Append(entity.LLMContext{
		Role:    entity.LLMRoleModel,
		Content: res,
	})
	out.LLMContext = llmContext

	return out, nil
}

func newHttpClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}

func (c *Client) updateToken() error {
	data := url.Values{}
	data.Set("scope", "GIGACHAT_API_CORP")

	req, err := http.NewRequest("POST", "https://ngw.devices.sberbank.ru:9443/api/v2/oauth", bytes.NewBufferString(data.Encode()))

	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+c.masterToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Accept-Charset", "utf-8")
	req.Header.Add("RqUID", uuid.New().String())

	res, err := newHttpClient().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	tokenResponse := &UpdateTokenResponse{}
	if err := json.Unmarshal(body, tokenResponse); err != nil {
		return err
	}

	if tokenResponse.Error.Code != 0 {
		return errors.New(tokenResponse.Error.Message)
	}

	c.accessToken = tokenResponse.AccessToken
	c.expiresAt = time.UnixMilli(tokenResponse.ExpiresAt)

	return nil
}

type Request struct {
	Model             string      `json:"model"`
	Messages          []AIMessage `json:"messages"`
	Stream            bool        `json:"stream"`
	RepetitionPenalty float32     `json:"repetition_penalty"`
	Temperature       float32     `json:"temperature"`
	TopP              float32     `json:"top_p"`
}

type AIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Response struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (c *Client) request(ctx context.Context, messages []AIMessage, temperature float32) (string, error) {
	AIRequest := Request{
		Model:             c.GigaChatModel,
		Messages:          messages,
		Stream:            false,
		RepetitionPenalty: 1,
		Temperature:       temperature,
		TopP:              0.4,
	}

	jsonAIRequest, err := json.Marshal(AIRequest)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://gigachat.devices.sberbank.ru/api/v1/chat/completions", bytes.NewBuffer(jsonAIRequest))

	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.accessToken)

	resp, err := newHttpClient().Do(req)
	if err != nil {
		return "", err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	respStruct := new(Response)

	if err := json.Unmarshal(b, respStruct); err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status code %d, error: %s", resp.StatusCode, respStruct.Message)
	}

	if len(respStruct.Choices) == 0 {
		return "", errors.New("no choices found")
	}

	return respStruct.Choices[0].Message.Content, nil
}
