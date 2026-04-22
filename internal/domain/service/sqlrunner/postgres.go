package sqlrunner

import (
	"SQLFactory/pkg/failure"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

func queryPostgres(ctx context.Context, connString, q string, maxRows int) (*QueryResult, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, failure.NewInternalError(err)
	}
	defer pool.Close()

	rows, err := pool.Query(ctx, q)
	if err != nil {
		return nil, failure.NewInternalError(err)
	}
	defer rows.Close()

	fds := rows.FieldDescriptions()
	cols := make([]string, 0, len(fds))
	for _, fd := range fds {
		cols = append(cols, string(fd.Name))
	}

	res := &QueryResult{Header: cols, Data: make([][]string, 0)}
	for rows.Next() {
		if maxRows > 0 && len(res.Data) >= maxRows {
			break
		}
		vals, err := rows.Values()
		if err != nil {
			return nil, failure.NewInternalError(err)
		}
		row := make([]string, 0, len(cols))
		for i := range cols {
			if i < len(vals) {
				row = append(row, normalizeDBValue(vals[i]))
				continue
			}
			row = append(row, "")
		}
		res.Data = append(res.Data, row)
	}
	if err := rows.Err(); err != nil {
		return nil, failure.NewInternalError(err)
	}
	return res, nil
}

