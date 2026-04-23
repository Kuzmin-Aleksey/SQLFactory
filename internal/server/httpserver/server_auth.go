package httpserver

import (
	"SQLFactory/internal/domain/entity"
	"SQLFactory/internal/domain/service/auth"
	"SQLFactory/pkg/contextx"
	"SQLFactory/pkg/failure"
	"errors"
	"net/http"
	"strings"
)

type AuthServer struct {
	service *auth.Service
}

func NewAuthServer(s *auth.Service) AuthServer {
	return AuthServer{
		service: s,
	}
}

func (s *AuthServer) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	email := r.FormValue("email")
	if email == "" {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(errors.New("email is required")))
		return
	}
	password := r.FormValue("password")
	if password == "" {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(errors.New("password is required")))
		return
	}
	name := r.FormValue("name")
	if name == "" {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(errors.New("name is required")))
		return
	}

	user := &entity.User{
		Email:    email,
		Password: password,
		Name:     name,
	}

	if err := s.service.Register(ctx, user); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *AuthServer) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	email := r.FormValue("email")
	if email == "" {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(errors.New("email is required")))
		return
	}
	password := r.FormValue("password")
	if password == "" {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(errors.New("password is required")))
		return
	}

	tokens, err := s.service.Login(ctx, email, password)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}

	writeJson(ctx, w, tokens, http.StatusOK)
}

func (s *AuthServer) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	refreshToken := r.FormValue("refresh_token")
	if refreshToken == "" {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(errors.New("refresh_token is required")))
		return
	}

	tokens, err := s.service.Refresh(ctx, refreshToken)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}

	writeJson(ctx, w, tokens, http.StatusOK)
}

func (s *AuthServer) MwAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if contextx.GetSkipAuth(ctx) {
			next(w, r)
			return
		}

		header := strings.Split(r.Header.Get("Authorization"), " ")
		if len(header) != 2 {
			writeAndLogErr(ctx, w, failure.NewInvalidRequestError(errors.New("invalid authorization header")))
			return
		}
		if header[0] != "Bearer" {
			writeAndLogErr(ctx, w, failure.NewInvalidRequestError(errors.New("invalid authorization header")))
			return
		}

		userId, err := s.service.Auth(ctx, header[1])
		if err != nil {
			writeAndLogErr(ctx, w, err)
			return
		}
		ctx = contextx.WithUserId(ctx, contextx.UserId(userId))
		r = r.WithContext(ctx)
		next(w, r)
	}
}
