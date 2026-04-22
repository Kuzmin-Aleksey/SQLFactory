package httpserver

import (
	"SQLFactory/internal/domain/service/aiquery"
	"SQLFactory/internal/domain/service/sqlrunner"
	"SQLFactory/pkg/failure"
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

type AIQueryService interface {
	Run(ctx context.Context, req aiquery.Request) (*aiquery.Response, error)
}

type AIQueryServer struct {
	service AIQueryService
}

func NewAIQueryServer(s AIQueryService) AIQueryServer {
	return AIQueryServer{service: s}
}

type AIQueryRequest struct {
	DBID       string                    `json:"db_id"`
	Text       string                    `json:"text"`
	Dict       map[string]string         `json:"dict"`
	Connection sqlrunner.ConnectionRequest `json:"connection"`
}

func (s *AIQueryServer) runAIQuery(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := AIQueryRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(err))
		return
	}
	if req.DBID == "" {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(errors.New("db_id is required")))
		return
	}
	if req.Text == "" {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(errors.New("text is required")))
		return
	}

	resp, err := s.service.Run(ctx, aiquery.Request{
		DBID:       req.DBID,
		Text:       req.Text,
		Dict:       req.Dict,
		Connection: req.Connection,
	})
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
	writeJson(ctx, w, resp, http.StatusOK)
}

