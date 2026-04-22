package aiquery

import (
	"SQLFactory/internal/domain/entity"
	"SQLFactory/internal/domain/service/sqlrunner"
	"SQLFactory/pkg/contextx"
	"SQLFactory/pkg/failure"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type LLM interface {
	GenerateIntent(ctx context.Context, input IntentInput) (IntentJSON, string, error)
	GenerateSQL(ctx context.Context, input SQLInput) (SQLJSON, string, error)
}

type HistorySaver interface {
	SaveItem(ctx context.Context, item *entity.HistoryItem) error
}

type Service struct {
	sqlrunner *sqlrunner.Service
	llm       LLM
	history   HistorySaver
}

func NewService(sqlRunner *sqlrunner.Service, llm LLM, history HistorySaver) *Service {
	return &Service{sqlrunner: sqlRunner, llm: llm, history: history}
}

type IntentInput struct {
	Text   string
	DBID   string
	Dict   map[string]string
	Schema *sqlrunner.DatabaseSchema
}

type SQLInput struct {
	Intent IntentJSON
	Schema *sqlrunner.DatabaseSchema
	DBID   string
}

func (s *Service) Run(ctx context.Context, req Request) (*Response, error) {
	if strings.TrimSpace(req.Text) == "" {
		return nil, failure.NewInvalidRequestError(errors.New("text is required"))
	}

	conn, err := s.sqlrunner.Connect(ctx, req.Connection)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	schemaCtx, cancelSchema := context.WithTimeout(ctx, s.sqlrunner.QueryTimeout)
	schema, err := conn.Schema(schemaCtx)
	cancelSchema()
	if err != nil {
		return nil, err
	}

	intent, _, err := s.llm.GenerateIntent(ctx, IntentInput{
		Text:   req.Text,
		DBID:   req.DBID,
		Dict:   req.Dict,
		Schema: schema,
	})
	if err != nil {
		return nil, err
	}

	sqlOut, _, err := s.llm.GenerateSQL(ctx, SQLInput{
		Intent: intent,
		Schema: schema,
		DBID:   req.DBID,
	})
	if err != nil {
		return nil, err
	}

	qctx, cancelQuery := context.WithTimeout(ctx, s.sqlrunner.QueryTimeout)
	data, err := conn.Query(qctx, sqlOut.SQL)
	cancelQuery()
	if err != nil {
		return nil, err
	}

	resp := &Response{
		SQL:              sqlOut.SQL,
		ExplanationSteps: sqlOut.ExplanationSteps,
		Chart:            sqlOut.Chart,
		Data:             data,
	}

	_ = s.saveHistory(ctx, req, resp)
	return resp, nil
}

func (s *Service) saveHistory(ctx context.Context, req Request, resp *Response) error {
	if s.history == nil {
		return nil
	}

	dataJSON, err := json.Marshal(resp.Data)
	if err != nil {
		return err
	}

	reasoning := strings.Join(resp.ExplanationSteps, "\n")
	title := req.Text
	if len(title) > 80 {
		title = title[:80]
	}

	item := &entity.HistoryItem{
		UserID:    int(contextx.GetUserId(ctx)),
		DB:        req.DBID,
		CreateAt:  time.Now(),
		Title:     title,
		Prompt:    req.Text,
		Query:     resp.SQL,
		Data:      string(dataJSON),
		Reasoning: reasoning,
	}

	return s.history.SaveItem(ctx, item)
}

