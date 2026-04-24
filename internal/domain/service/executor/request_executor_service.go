package executor

import (
	"SQLFactory/internal/domain/entity"
	"SQLFactory/internal/domain/service/sqlrunner"
	"SQLFactory/internal/domain/value"
	"SQLFactory/pkg/contextx"
	"SQLFactory/pkg/failure"
	"context"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type History interface {
	SaveItem(ctx context.Context, item *entity.HistoryItem) error
	GetItem(ctx context.Context, id int) (*entity.HistoryItem, error)
	UpdateTableData(ctx context.Context, id int, data value.JsonValue) error
}

type Dict interface {
	GetByDBMap(ctx context.Context, dbId string) (map[string]string, error)
}

type Templates interface {
	GetById(ctx context.Context, id int) (*entity.Template, error)
}

type LLMResponse struct {
	Title            string   `json:"title"`
	SQL              string   `json:"sql"`
	ExplanationSteps []string `json:"explanation_steps"`
	ChartType        string   `json:"chart_type"`
	NeedQuery        bool     `json:"need_query"`
	LLMContext       any
}

type LLM interface {
	GenerateSQL(ctx context.Context, request string, dict map[string]string, schema any, dbType string) (*LLMResponse, error)
	GenerateSQLSecond(ctx context.Context, LLMContext any, data any) (*LLMResponse, error)
}

type Service struct {
	history   History
	templates Templates
	dict      Dict
	sqlRunner *sqlrunner.Service
	llm       LLM
}

func NewService(history History, templates Templates, dict Dict, sqlRunner *sqlrunner.Service, llm LLM) *Service {
	return &Service{
		history:   history,
		templates: templates,
		dict:      dict,
		sqlRunner: sqlRunner,
		llm:       llm,
	}
}

type LLMExecuteResult struct {
	Query        string          `json:"query"`
	Title        string          `json:"title"`
	TableData    value.JsonValue `json:"table_data"`
	ChartType    string          `json:"chart_type"`
	Reasoning    value.JsonValue `json:"reasoning"`
	HistoryId    int             `json:"history_id"`
	ExecuteError string          `json:"execute_error,omitempty"`
}

type ExecuteResult struct {
	TableData    value.JsonValue `json:"table_data"`
	ExecuteError string          `json:"execute_error,omitempty"`
}

func (s *Service) Connect(ctx context.Context, connCfg sqlrunner.ConnectionRequest) (string, error) {
	const op = "request_template_service/Connect"

	conn, err := s.sqlRunner.Connect(ctx, connCfg)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	conn.Close()

	dbId := hashConnConfig(connCfg)
	return dbId, nil
}

func (s *Service) ExecuteUserRequest(ctx context.Context, connConfig sqlrunner.ConnectionRequest, prompt string) (*LLMExecuteResult, error) {
	const op = "request_template_service/ExecuteUserRequest"
	l := contextx.GetLoggerOrDefault(ctx)

	conn, err := s.sqlRunner.Connect(ctx, connConfig)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer conn.Close()

	schema, err := conn.Schema(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	dbId := hashConnConfig(connConfig)
	dict, err := s.dict.GetByDBMap(ctx, dbId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	llmResp, err := s.llm.GenerateSQL(ctx, prompt, dict, schema, connConfig.DBType)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	l.Debug("generated llm response", "resp", llmResp)

	reasoningJson, _ := json.Marshal(llmResp.ExplanationSteps)

	var executeError string

	var result *sqlrunner.QueryResult

	if llmResp.NeedQuery {
		var resultForLLM any
		resultForLLM, err = conn.Query(ctx, llmResp.SQL)
		if err != nil {
			resultForLLM = err.Error()
		}
		l.Debug("result for llm", "result", resultForLLM)

		llmResp, err = s.llm.GenerateSQLSecond(ctx, llmResp.LLMContext, resultForLLM)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		l.Debug("generated second llm response", "resp", llmResp)
	}

	result, err = conn.Query(ctx, llmResp.SQL)
	if err != nil {
		if dbError := new(failure.ExternalDBError); errors.As(err, dbError) {
			executeError = dbError.Error()
		} else {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		return &LLMExecuteResult{
			Query:        llmResp.SQL,
			Title:        llmResp.Title,
			Reasoning:    value.JsonValue(reasoningJson),
			ExecuteError: executeError,
		}, nil

	} else {
		jsonResult, _ := json.Marshal(result)
		tableData := value.JsonValue(jsonResult)

		historyItem := &entity.HistoryItem{
			UserID:    int(contextx.GetUserId(ctx)),
			DB:        hashConnConfig(connConfig),
			CreateAt:  time.Now(),
			Title:     llmResp.Title,
			Prompt:    prompt,
			Query:     llmResp.SQL,
			TableData: tableData,
			ChartType: llmResp.ChartType,
			Reasoning: value.JsonValue(reasoningJson),
		}

		if err := s.history.SaveItem(ctx, historyItem); err != nil {
			contextx.GetLoggerOrDefault(ctx).Error("failed save history", "err", err, "HistoryItem", historyItem)
		}

		return &LLMExecuteResult{
			Query:     llmResp.SQL,
			Title:     llmResp.Title,
			TableData: tableData,
			ChartType: llmResp.ChartType,
			Reasoning: value.JsonValue(reasoningJson),
			HistoryId: historyItem.Id,
		}, nil
	}
}

func (s *Service) ExecuteTemplate(ctx context.Context, connConfig sqlrunner.ConnectionRequest, templateId int) (*ExecuteResult, error) {
	const op = "request_template_service/ExecuteTemplate"
	tmpl, err := s.templates.GetById(ctx, templateId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	result, err := s.sqlRunner.Query(ctx, sqlrunner.QueryRequest{
		SQL:               tmpl.Query,
		ConnectionRequest: connConfig,
	})
	if err != nil {
		if dbError := new(failure.ExternalDBError); errors.As(err, dbError) {
			return &ExecuteResult{
				ExecuteError: dbError.Error(),
			}, nil
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	resultJson, _ := json.Marshal(result)
	return &ExecuteResult{
		TableData: value.JsonValue(resultJson),
	}, nil

}

func (s *Service) ExecuteHistoryItem(ctx context.Context, connConfig sqlrunner.ConnectionRequest, itemId int) (*ExecuteResult, error) {
	const op = "request_template_service/ExecuteHistoryItem"
	historyItem, err := s.history.GetItem(ctx, itemId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	result, err := s.sqlRunner.Query(ctx, sqlrunner.QueryRequest{
		SQL:               historyItem.Query,
		ConnectionRequest: connConfig,
	})
	if err != nil {
		if dbError := new(failure.ExternalDBError); errors.As(err, dbError) {
			return &ExecuteResult{
				ExecuteError: dbError.Error(),
			}, nil
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	resultJson, _ := json.Marshal(result)
	tableData := value.JsonValue(resultJson)

	if err := s.history.UpdateTableData(ctx, itemId, tableData); err != nil {
		contextx.GetLoggerOrDefault(ctx).Error("failed update table data", "err", err, "item_id", itemId, "table_data", tableData)
	}

	return &ExecuteResult{
		TableData: tableData,
	}, nil
}

func hashConnConfig(connCfg sqlrunner.ConnectionRequest) string {
	hash := sha512.New()
	hash.Write([]byte(connCfg.DBType))
	hash.Write([]byte(connCfg.Host))
	binary.Write(hash, binary.BigEndian, connCfg.Port)
	hash.Write([]byte(connCfg.Database))
	return hex.EncodeToString(hash.Sum(nil))
}
