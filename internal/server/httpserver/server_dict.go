package httpserver

import (
	"SQLFactory/internal/domain/entity"
	"SQLFactory/pkg/failure"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
)

type DictService interface {
	Add(ctx context.Context, dictItem *entity.DictItem) error
	Update(ctx context.Context, item *entity.DictItem) error
	GetByDB(ctx context.Context, dbId string) (map[string]string, error)
	Delete(ctx context.Context, id int) error
}

type DictServer struct {
	service DictService
}

func NewDictServer(s DictService) DictServer {
	return DictServer{s}
}

func (s *DictServer) addDictItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	item := new(entity.DictItem)
	if err := json.NewDecoder(r.Body).Decode(item); err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(err))
		return
	}
	if err := s.service.Add(ctx, item); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
	writeJson(ctx, w, newIdResponse(item.Id), http.StatusOK)
}

func (s *DictServer) updateDictItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	item := new(entity.DictItem)
	if err := json.NewDecoder(r.Body).Decode(item); err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(err))
		return
	}
	if err := s.service.Update(ctx, item); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *DictServer) getDBDict(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dbId := r.FormValue("db")
	dict, err := s.service.GetByDB(ctx, dbId)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
	writeJson(ctx, w, dict, http.StatusOK)
}

func (s *DictServer) deleteDictItem(w http.ResponseWriter, r *http.Request) {
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
