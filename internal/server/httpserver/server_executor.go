package httpserver

import (
	"SQLFactory/internal/domain/service/executor"
	"SQLFactory/internal/domain/service/sqlrunner"
	"SQLFactory/pkg/failure"
	"errors"
	"net/http"
	"strconv"
)

type ExecutorServer struct {
	service *executor.Service
}

func NewExecutorServer(service *executor.Service) ExecutorServer {
	return ExecutorServer{
		service: service,
	}
}

func (s *ExecutorServer) executeUserPrompt(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var connCfg sqlrunner.ConnectionRequest
	var err error

	connCfg.Host = r.FormValue("host")
	connCfg.Port, err = strconv.Atoi(r.FormValue("port"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(err))
		return
	}
	connCfg.User = r.FormValue("user")
	connCfg.Password = r.FormValue("password")
	connCfg.Database = r.FormValue("database")
	connCfg.DBType = r.FormValue("db_type")

	prompt := r.FormValue("prompt")
	if prompt == "" {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(errors.New("prompt is required")))
	}

	result, err := s.service.ExecuteUserRequest(ctx, connCfg, prompt)
	if err != nil {
		var errSQlExternal failure.ExternalDBError
		if errors.As(err, &errSQlExternal) {
			writeJson(ctx, w, errorResponse{
				Error: errSQlExternal.Error(),
			}, http.StatusOK)
		} else {
			writeAndLogErr(ctx, w, err)
		}
	}
	writeJson(ctx, w, result, http.StatusOK)
}

func (s *ExecutorServer) executeTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var connCfg sqlrunner.ConnectionRequest
	var err error

	connCfg.Host = r.FormValue("host")
	connCfg.Port, err = strconv.Atoi(r.FormValue("port"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(err))
		return
	}
	connCfg.User = r.FormValue("user")
	connCfg.Password = r.FormValue("password")
	connCfg.Database = r.FormValue("database")
	connCfg.DBType = r.FormValue("db_type")

	templateId, err := strconv.Atoi(r.FormValue("template_id"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(err))
		return
	}

	result, err := s.service.ExecuteTemplate(ctx, connCfg, templateId)
	if err != nil {
		var errSQlExternal failure.ExternalDBError
		if errors.As(err, &errSQlExternal) {
			writeJson(ctx, w, errorResponse{
				Error: errSQlExternal.Error(),
			}, http.StatusOK)
		} else {
			writeAndLogErr(ctx, w, err)
		}
	}

	writeJson(ctx, w, result, http.StatusOK)
}

func (s *ExecutorServer) executeHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var connCfg sqlrunner.ConnectionRequest
	var err error

	connCfg.Host = r.FormValue("host")
	connCfg.Port, err = strconv.Atoi(r.FormValue("port"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(err))
		return
	}
	connCfg.User = r.FormValue("user")
	connCfg.Password = r.FormValue("password")
	connCfg.Database = r.FormValue("database")
	connCfg.DBType = r.FormValue("db_type")

	historyItemId, err := strconv.Atoi(r.FormValue("item_id"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(err))
		return
	}

	result, err := s.service.ExecuteHistoryItem(ctx, connCfg, historyItemId)
	if err != nil {
		var errSQlExternal failure.ExternalDBError
		if errors.As(err, &errSQlExternal) {
			writeJson(ctx, w, errorResponse{
				Error: errSQlExternal.Error(),
			}, http.StatusOK)
		} else {
			writeAndLogErr(ctx, w, err)
		}
	}

	writeJson(ctx, w, result, http.StatusOK)
}
