package sqlrunner

import (
	"SQLFactory/pkg/failure"
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type Service struct {
	QueryTimeout time.Duration
	MaxRows      int
}

func NewService() *Service {
	return &Service{
		QueryTimeout: 10 * time.Second,
		MaxRows:      1000,
	}
}

func (s *Service) Query(ctx context.Context, req QueryRequest) (*QueryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, s.QueryTimeout)
	defer cancel()

	dbType := strings.ToLower(strings.TrimSpace(req.DBType))
	switch dbType {
	case "mysql":
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true",
			req.User, req.Password, req.Host, req.Port, req.Database,
		)
		return queryMySQL(ctx, dsn, req.SQL, s.MaxRows)
	case "postgres", "postgresql":
		u := &url.URL{
			Scheme: "postgresql",
			User:   url.UserPassword(req.User, req.Password),
			Host:   fmt.Sprintf("%s:%d", req.Host, req.Port),
			Path:   "/" + req.Database,
		}
		return queryPostgres(ctx, u.String(), req.SQL, s.MaxRows)
	default:
		return nil, failure.ErrUnsupportedDBType
	}
}
