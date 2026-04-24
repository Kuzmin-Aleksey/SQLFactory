package httpserver

import (
	"SQLFactory/internal/domain/entity"
	"SQLFactory/pkg/failure"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
)

type TemplatesService interface {
	SaveTemplate(ctx context.Context, template *entity.Template) error
	UpdateTemplate(ctx context.Context, template *entity.Template) error
	GetById(ctx context.Context, id int) (*entity.Template, error)
	GetDBTemplates(ctx context.Context, dbId string) ([]entity.Template, error)
	DeleteTemplate(ctx context.Context, id int) error
}

type TemplatesServer struct {
	service TemplatesService
}

func NewTemplatesServer(s TemplatesService) TemplatesServer {
	return TemplatesServer{s}
}

func (s *TemplatesServer) saveTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	template := new(entity.Template)
	if err := json.NewDecoder(r.Body).Decode(template); err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(err))
		return
	}
	if err := s.service.SaveTemplate(ctx, template); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
	writeJson(ctx, w, newIdResponse(template.Id), http.StatusOK)
}

func (s *TemplatesServer) updateTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	template := new(entity.Template)
	if err := json.NewDecoder(r.Body).Decode(template); err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(err))
		return
	}
	if err := s.service.UpdateTemplate(ctx, template); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *TemplatesServer) getTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(err))
		return
	}
	template, err := s.service.GetById(ctx, id)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
	writeJson(ctx, w, template, http.StatusOK)
}

func (s *TemplatesServer) getDBTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dbId := r.FormValue("db")
	templates, err := s.service.GetDBTemplates(ctx, dbId)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
	writeJson(ctx, w, templates, http.StatusOK)
}

func (s *TemplatesServer) deleteTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(err))
		return
	}
	if err := s.service.DeleteTemplate(ctx, id); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}
