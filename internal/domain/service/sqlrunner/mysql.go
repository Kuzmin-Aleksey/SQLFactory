package sqlrunner

import (
	"SQLFactory/pkg/failure"
	"context"
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

func queryMySQL(ctx context.Context, dsn, q string, maxRows int) (*QueryResult, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, failure.NewInternalError(err)
	}
	defer db.Close()

	db.SetConnMaxLifetime(2 * time.Minute)
	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(2)

	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, failure.NewInternalError(err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, failure.NewInternalError(err)
	}

	res := &QueryResult{Header: cols, Data: make([][]string, 0)}

	for rows.Next() {
		if maxRows > 0 && len(res.Data) >= maxRows {
			break
		}
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, failure.NewInternalError(err)
		}
		row := make([]string, 0, len(cols))
		for i := range cols {
			row = append(row, normalizeDBValue(values[i]))
		}
		res.Data = append(res.Data, row)
	}
	if err := rows.Err(); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		return nil, failure.NewInternalError(err)
	}

	return res, nil
}

