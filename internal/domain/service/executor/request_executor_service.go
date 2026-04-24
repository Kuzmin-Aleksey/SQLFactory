package executor

import (
	"SQLFactory/internal/domain/entity"
	"SQLFactory/internal/domain/service/sqlrunner"
	"SQLFactory/internal/domain/value"
	"SQLFactory/internal/infrastructure/llm"
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

type LLMContextRepo interface {
	Save(ctx context.Context, llmContext *entity.LLMContext) error
	GetFullContextByHistoryId(ctx context.Context, historyId int) ([]entity.LLMContext, error)
}

type LLM interface {
	GenerateSQL(ctx context.Context, llmContext llm.Context, request string, dict map[string]string, schema any, dbType string) (*llm.Response, error)
	GenerateSQLSecond(ctx context.Context, llmContext llm.Context, data any) (*llm.Response, error)
	RegenerateSQL(ctx context.Context, llmContext llm.Context, request string) (*llm.Response, error)
}

type Service struct {
	history        History
	templates      Templates
	dict           Dict
	llmContextRepo LLMContextRepo
	sqlRunner      *sqlrunner.Service
	llm            LLM
}

func NewService(history History, templates Templates, dict Dict, llmContextRepo LLMContextRepo, sqlRunner *sqlrunner.Service, llm LLM) *Service {
	return &Service{
		history:        history,
		templates:      templates,
		dict:           dict,
		llmContextRepo: llmContextRepo,
		sqlRunner:      sqlRunner,
		llm:            llm,
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
	Query        string          `json:"query"`
	Title        string          `json:"title"`
	TableData    value.JsonValue `json:"table_data"`
	ChartType    string          `json:"chart_type"`
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

func (s *Service) ExecuteUserRequest(ctx context.Context, connConfig sqlrunner.ConnectionRequest, prompt string, previousHistoryId *int) (*LLMExecuteResult, error) {
	const op = "request_template_service/ExecuteUserRequest"
	l := contextx.GetLoggerOrDefault(ctx)

	conn, err := s.sqlRunner.Connect(ctx, connConfig)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer conn.Close()

	var llmResp *llm.Response

	var llmContext llm.Context
	if previousHistoryId != nil {
		llmContext, err = s.llmContextRepo.GetFullContextByHistoryId(ctx, *previousHistoryId)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		l.Debug("using saved context", "context", llmContext)

		llmResp, err = s.llm.RegenerateSQL(ctx, llmContext, prompt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		l.Debug("generated llm response", "resp", llmResp)

	} else {
		schema, err := conn.Schema(ctx)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		dbId := hashConnConfig(connConfig)
		dict, err := s.dict.GetByDBMap(ctx, dbId)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		llmResp, err = s.llm.GenerateSQL(ctx, llmContext, prompt, dict, schema, connConfig.DBType)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		l.Debug("generated llm response", "resp", llmResp)
	}

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

	reasoningJson, _ := json.Marshal(llmResp.ExplanationSteps)

	result, err = conn.Query(ctx, llmResp.SQL)
	jsonResult, _ := json.Marshal(result)
	tableData := value.JsonValue(jsonResult)

	var executeError string
	if err != nil {
		if dbError := new(failure.ExternalDBError); errors.As(err, dbError) {
			executeError = dbError.Error()
		} else {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	historyItem := &entity.HistoryItem{
		UserID:     int(contextx.GetUserId(ctx)),
		PreviousId: previousHistoryId,
		DB:         hashConnConfig(connConfig),
		CreateAt:   time.Now(),
		Title:      llmResp.Title,
		Prompt:     prompt,
		Query:      llmResp.SQL,
		TableData:  tableData,
		ChartType:  llmResp.ChartType,
		Reasoning:  value.JsonValue(reasoningJson),
	}

	if err := s.history.SaveItem(ctx, historyItem); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if err := s.saveLlmContext(ctx, historyItem.Id, llmResp.LLMContext); err != nil {
		contextx.GetLoggerOrDefault(ctx).Error("failed save llm context", "err", err, "HistoryItem", historyItem)
	}

	return &LLMExecuteResult{
		Query:        llmResp.SQL,
		Title:        llmResp.Title,
		TableData:    tableData,
		ChartType:    llmResp.ChartType,
		Reasoning:    value.JsonValue(reasoningJson),
		HistoryId:    historyItem.Id,
		ExecuteError: executeError,
	}, nil
}

func (s *Service) saveLlmContext(ctx context.Context, historyId int, llmContext llm.Context) error {
	const op = "request_template_service/SaveLlmContext"
	for i, item := range llmContext {
		if item.Id == 0 {
			item.HistoryId = historyId
			if err := s.llmContextRepo.Save(ctx, &item); err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
			if i != len(llmContext)-1 {
				llmContext[i+1].PreviousId = &item.Id
			}
		}
	}
	return nil
}

func (s *Service) ExecuteTemplate(ctx context.Context, connConfig sqlrunner.ConnectionRequest, templateId int) (*ExecuteResult, error) {
	const op = "request_template_service/ExecuteTemplate"
	tmpl, err := s.templates.GetById(ctx, templateId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	executeResult := &ExecuteResult{
		Query:     tmpl.Query,
		Title:     tmpl.Title,
		ChartType: tmpl.ChartType,
	}

	result, err := s.sqlRunner.Query(ctx, sqlrunner.QueryRequest{
		SQL:               tmpl.Query,
		ConnectionRequest: connConfig,
	})
	if err != nil {
		if dbError := new(failure.ExternalDBError); errors.As(err, dbError) {
			executeResult.ExecuteError = dbError.Error()
			return executeResult, nil
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	resultJson, _ := json.Marshal(result)
	executeResult.TableData = value.JsonValue(resultJson)

	return executeResult, nil

}

func (s *Service) ExecuteHistoryItem(ctx context.Context, connConfig sqlrunner.ConnectionRequest, itemId int) (*ExecuteResult, error) {
	const op = "request_template_service/ExecuteHistoryItem"
	historyItem, err := s.history.GetItem(ctx, itemId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	executeResult := &ExecuteResult{
		Query:     historyItem.Query,
		Title:     historyItem.Title,
		ChartType: historyItem.ChartType,
	}

	result, err := s.sqlRunner.Query(ctx, sqlrunner.QueryRequest{
		SQL:               historyItem.Query,
		ConnectionRequest: connConfig,
	})
	if err != nil {
		if dbError := new(failure.ExternalDBError); errors.As(err, dbError) {
			executeResult.ExecuteError = dbError.Error()
			return executeResult, nil
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	resultJson, _ := json.Marshal(result)
	tableData := value.JsonValue(resultJson)

	if err := s.history.UpdateTableData(ctx, itemId, tableData); err != nil {
		contextx.GetLoggerOrDefault(ctx).Error("failed update table data", "err", err, "item_id", itemId, "table_data", tableData)
	}

	executeResult.TableData = tableData

	return executeResult, nil
}

func hashConnConfig(connCfg sqlrunner.ConnectionRequest) string {
	hash := sha512.New()
	hash.Write([]byte(connCfg.DBType))
	hash.Write([]byte(connCfg.Host))
	binary.Write(hash, binary.BigEndian, connCfg.Port)
	hash.Write([]byte(connCfg.Database))
	return hex.EncodeToString(hash.Sum(nil))
}
