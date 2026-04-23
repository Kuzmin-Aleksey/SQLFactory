package sqlrunner

import (
	"SQLFactory/pkg/failure"
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresConn struct {
	pool    *pgxpool.Pool
	dbName  string
	maxRows int
}

func (c *postgresConn) Query(ctx context.Context, q string) (*QueryResult, error) {
	return queryPostgresWithPool(ctx, c.pool, q, c.maxRows)
}

func (c *postgresConn) Schema(ctx context.Context) (*DatabaseSchema, error) {
	return schemaPostgres(ctx, c.pool, c.dbName)
}

func (c *postgresConn) Close() error {
	c.pool.Close()
	return nil
}

func connectPostgres(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, failure.NewExternalDBError(err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, failure.NewExternalDBError(err)
	}
	return pool, nil
}

func queryPostgresWithPool(ctx context.Context, pool *pgxpool.Pool, q string, maxRows int) (*QueryResult, error) {
	rows, err := pool.Query(ctx, q)
	if err != nil {
		return nil, failure.NewExternalDBError(err)
	}
	defer rows.Close()

	fds := rows.FieldDescriptions()
	cols := make([]string, 0, len(fds))
	for _, fd := range fds {
		cols = append(cols, string(fd.Name))
	}

	res := &QueryResult{Header: cols}
	res.Data = make([][]string, 0)
	for rows.Next() {
		if maxRows > 0 && len(res.Data) >= maxRows {
			break
		}

		vals, err := rows.Values()
		if err != nil {
			return nil, failure.NewExternalDBError(err)
		}

		row := make([]string, len(cols))
		for i := range row {
			if i >= len(vals) {
				break
			}
			row[i] = normalizeDBValue(vals[i])
		}
		res.Data = append(res.Data, row)
	}

	if err := rows.Err(); err != nil {
		return nil, failure.NewExternalDBError(err)
	}
	return res, nil
}

func schemaPostgres(ctx context.Context, pool *pgxpool.Pool, dbName string) (*DatabaseSchema, error) {
	const colsSQL = `
SELECT c.table_schema, c.table_name, c.column_name, c.data_type, c.is_nullable, c.column_default
FROM information_schema.columns c
JOIN information_schema.tables t
  ON t.table_schema = c.table_schema AND t.table_name = c.table_name
WHERE t.table_type = 'BASE TABLE'
  AND c.table_schema NOT IN ('pg_catalog', 'information_schema')
ORDER BY c.table_schema, c.table_name, c.ordinal_position
`
	rows, err := pool.Query(ctx, colsSQL)
	if err != nil {
		return nil, failure.NewExternalDBError(err)
	}
	defer rows.Close()

	type tableKey struct {
		schema string
		name   string
	}
	tables := make(map[tableKey]*TableSchema)
	order := make([]tableKey, 0)

	for rows.Next() {
		var schemaName, tableName, colName, dataType, isNullable string
		var colDefault *string

		if err := rows.Scan(&schemaName, &tableName, &colName, &dataType, &isNullable, &colDefault); err != nil {
			return nil, failure.NewExternalDBError(err)
		}

		k := tableKey{schema: schemaName, name: tableName}
		t, ok := tables[k]
		if !ok {
			t = &TableSchema{Schema: schemaName, Name: tableName}
			tables[k] = t
			order = append(order, k)
		}

		c := ColumnSchema{
			Name:     colName,
			DataType: dataType,
			Nullable: strings.EqualFold(isNullable, "YES"),
		}
		if colDefault != nil {
			c.Default = *colDefault
		}
		t.Columns = append(t.Columns, c)
	}
	if err := rows.Err(); err != nil {
		return nil, failure.NewExternalDBError(err)
	}

	const fkSQL = `
SELECT
  tc.table_schema,
  tc.table_name,
  kcu.column_name,
  tc.constraint_name,
  ccu.table_schema AS foreign_table_schema,
  ccu.table_name AS foreign_table_name,
  ccu.column_name AS foreign_column_name
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
  ON tc.constraint_name = kcu.constraint_name AND tc.table_schema = kcu.table_schema
JOIN information_schema.constraint_column_usage ccu
  ON ccu.constraint_name = tc.constraint_name AND ccu.table_schema = tc.table_schema
WHERE tc.constraint_type = 'FOREIGN KEY'
  AND tc.table_schema NOT IN ('pg_catalog', 'information_schema')
ORDER BY tc.table_schema, tc.table_name, kcu.column_name
`
	fkRows, err := pool.Query(ctx, fkSQL)
	if err != nil {
		return nil, failure.NewExternalDBError(err)
	}
	defer fkRows.Close()

	for fkRows.Next() {
		var schemaName, tableName, colName, constraintName, refSchema, refTable, refCol string
		if err := fkRows.Scan(&schemaName, &tableName, &colName, &constraintName, &refSchema, &refTable, &refCol); err != nil {
			return nil, failure.NewExternalDBError(err)
		}

		k := tableKey{schema: schemaName, name: tableName}
		t, ok := tables[k]
		if !ok {
			t = &TableSchema{Schema: schemaName, Name: tableName}
			tables[k] = t
			order = append(order, k)
		}
		t.ForeignKeys = append(t.ForeignKeys, ForeignKeySchema{
			ConstraintName: constraintName,
			Column:         colName,
			RefSchema:      refSchema,
			RefTable:       refTable,
			RefColumn:      refCol,
		})
	}
	if err := fkRows.Err(); err != nil {
		return nil, failure.NewExternalDBError(err)
	}

	out := &DatabaseSchema{Database: dbName, Tables: make([]TableSchema, 0, len(order))}
	for _, k := range order {
		if t := tables[k]; t != nil {
			out.Tables = append(out.Tables, *t)
		}
	}
	return out, nil
}
