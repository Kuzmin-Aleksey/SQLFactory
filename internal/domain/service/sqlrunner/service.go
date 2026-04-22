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

type Conn interface {
	Query(ctx context.Context, sql string) (*QueryResult, error)
	Schema(ctx context.Context) (*DatabaseSchema, error)
	Close() error
}

func NewService() *Service {
	return &Service{
		QueryTimeout: 10 * time.Second,
		MaxRows:      1000,
	}
}

func (s *Service) Query(ctx context.Context, req QueryRequest) (*QueryResult, error) {
	conn, err := s.Connect(ctx, req.ConnectionRequest)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	qctx, cancel := context.WithTimeout(ctx, s.QueryTimeout)
	defer cancel()

	return conn.Query(qctx, req.SQL)
}

func (s *Service) Connect(ctx context.Context, req ConnectionRequest) (Conn, error) {
	dbType := strings.ToLower(strings.TrimSpace(req.DBType))
	switch dbType {
	case "mysql":
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true",
			req.User, req.Password, req.Host, req.Port, req.Database,
		)
		return connectMySQLConn(ctx, dsn, req.Database, s.MaxRows)
	case "postgres", "postgresql":
		u := &url.URL{
			Scheme: "postgresql",
			User:   url.UserPassword(req.User, req.Password),
			Host:   fmt.Sprintf("%s:%d", req.Host, req.Port),
			Path:   "/" + req.Database,
		}
		return connectPostgresConn(ctx, u.String(), req.Database, s.MaxRows)
	default:
		return nil, failure.ErrUnsupportedDBType
	}
}

func connectMySQLConn(ctx context.Context, dsn string, dbName string, maxRows int) (Conn, error) {
	db, err := connectMySQL(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &mysqlConn{db: db, dbName: dbName, maxRows: maxRows}, nil
}

func connectPostgresConn(ctx context.Context, connString string, dbName string, maxRows int) (Conn, error) {
	pool, err := connectPostgres(ctx, connString)
	if err != nil {
		return nil, err
	}
	return &postgresConn{pool: pool, dbName: dbName, maxRows: maxRows}, nil
}
