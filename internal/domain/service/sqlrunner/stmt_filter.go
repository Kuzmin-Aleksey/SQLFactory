package sqlrunner

import (
	"SQLFactory/pkg/contextx"
	"SQLFactory/pkg/failure"
	"context"
	"errors"
	"github.com/oarkflow/sqlparser"
)

type stmtFilter struct {
	Conn
}

func withStmtFilter(conn Conn) Conn {
	return &stmtFilter{conn}
}

var ErrInvalidSQL = errors.New("sql query not SELECT")

func (s stmtFilter) Query(ctx context.Context, sql string) (*QueryResult, error) {
	stmt, err := sqlparser.ParseStatement(sql)
	if err != nil {
		// будем считать что это плохой парсер, а не запрос
		contextx.GetLoggerOrDefault(ctx).Error("failed to parse sql", "sql", sql, "err", err)
		return s.Conn.Query(ctx, sql)
	}

	if _, ok := stmt.(*sqlparser.SelectStmt); !ok {
		return nil, failure.NewExternalDBError(ErrInvalidSQL)
	}
	return s.Conn.Query(ctx, sql)
}
