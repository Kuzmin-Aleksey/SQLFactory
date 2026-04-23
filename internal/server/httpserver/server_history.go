package httpserver

import (
	"SQLFactory/internal/domain/entity"
	"SQLFactory/pkg/failure"
	"context"
	"net/http"
	"strconv"
)

type HistoryService interface {
	GetByDB(ctx context.Context, db string) ([]entity.HistoryItem, error)
	GetItem(ctx context.Context, id int) (*entity.HistoryItem, error)
	Delete(ctx context.Context, id int) error
}

type HistoryServer struct {
	service HistoryService
}

func NewHistoryServer(s HistoryService) HistoryServer {
	return HistoryServer{s}
}

func (s *HistoryServer) getDBHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := r.FormValue("db")
	history, err := s.service.GetByDB(ctx, db)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
	writeJson(ctx, w, history, http.StatusOK)
}

func (s *HistoryServer) getHistoryItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(err))
		return
	}

	item, err := s.service.GetItem(ctx, id)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
	writeJson(ctx, w, item, http.StatusOK)
}

func (s *HistoryServer) deleteHistoryItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(err))
		return
	}
	if err := s.service.Delete(ctx, id); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}
