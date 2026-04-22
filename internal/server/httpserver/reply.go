package httpserver

import (
	"SQLFactory/pkg/contextx"
	"SQLFactory/pkg/errcodes"
	"SQLFactory/pkg/failure"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeAndLogErr(ctx context.Context, w http.ResponseWriter, err error) {
	errCode, statusCode := getCodeFromError(err)
	writeJson(ctx, w, errorResponse{Error: errCode.String()}, statusCode)

	l := contextx.GetLoggerOrDefault(ctx)
	l.LogAttrs(ctx, slog.LevelError, "error handling request", slog.String("err", err.Error()))
}

func writeJson(ctx context.Context, w http.ResponseWriter, v any, status int) {
	l := contextx.GetLoggerOrDefault(ctx)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		l.LogAttrs(ctx, slog.LevelError, "json encode error", slog.String("err", err.Error()))
	}
}

type respId struct {
	Id any `json:"id"`
}

func newIdResponse(id any) respId {
	return respId{id}
}

func getCodeFromError(err error) (errcodes.Code, int) {
	switch {
	case failure.IsInternalError(err):
		return errcodes.Internal, http.StatusInternalServerError
	case failure.IsNotFoundError(err):
		return errcodes.NotFound, http.StatusNotFound
	case failure.IsInvalidRequestError(err):
		return errcodes.Validation, http.StatusBadRequest
	case failure.IsNotUnauthenticatedError(err):
		return errcodes.Unauthorized, http.StatusUnauthorized
	case failure.IsAlreadyExistsError(err):
		return errcodes.AlreadyExists, http.StatusConflict
	default:
		return errcodes.Internal, http.StatusInternalServerError
	}
}
