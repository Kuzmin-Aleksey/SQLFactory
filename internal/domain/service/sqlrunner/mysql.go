package sqlrunner

import (
	"SQLFactory/pkg/failure"
	"context"
	"database/sql"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type mysqlConn struct {
	db      *sql.DB
	dbName  string
	maxRows int
}

func (c *mysqlConn) Query(ctx context.Context, q string) (*QueryResult, error) {
	return queryMySQLWithDB(ctx, c.db, q, c.maxRows)
}

func (c *mysqlConn) Schema(ctx context.Context) (*DatabaseSchema, error) {
	return schemaMySQL(ctx, c.db, c.dbName)
}

func (c *mysqlConn) Close() error { return c.db.Close() }

func connectMySQL(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, failure.NewInternalError(err)
	}
	_ = ctx
	return db, nil
}

func queryMySQLWithDB(ctx context.Context, db *sql.DB, q string, maxRows int) (*QueryResult, error) {
	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, failure.NewInternalError(err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, failure.NewInternalError(err)
	}

	res := &QueryResult{Header: cols}
	for rows.Next() {
		if maxRows > 0 && len(res.Data) >= maxRows {
			break
		}

		values := make([]any, len(cols))
		scanArgs := make([]any, len(cols))
		for i := range scanArgs {
			scanArgs[i] = &values[i]
		}
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, failure.NewInternalError(err)
		}

		row := make([]string, len(cols))
		for i := range row {
			row[i] = NormalizeDBValue(values[i])
		}
		res.Data = append(res.Data, row)
	}

	if err := rows.Err(); err != nil {
		return nil, failure.NewInternalError(err)
	}

	return res, nil
}

func schemaMySQL(ctx context.Context, db *sql.DB, dbName string) (*DatabaseSchema, error) {
	// Columns
	const colsSQL = `
SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT, EXTRA, COLUMN_KEY
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_SCHEMA = ?
ORDER BY TABLE_NAME, ORDINAL_POSITION
`

	rows, err := db.QueryContext(ctx, colsSQL, dbName)
	if err != nil {
		return nil, failure.NewInternalError(err)
	}
	defer rows.Close()

	type tableKey struct {
		schema string
		name   string
	}

	tables := make(map[tableKey]*TableSchema)
	order := make([]tableKey, 0)

	for rows.Next() {
		var tableName, colName, dataType, isNullable string
		var colDefault sql.NullString
		var extra, colKey sql.NullString

		if err := rows.Scan(&tableName, &colName, &dataType, &isNullable, &colDefault, &extra, &colKey); err != nil {
			return nil, failure.NewInternalError(err)
		}

		k := tableKey{schema: "", name: tableName}
		t, ok := tables[k]
		if !ok {
			t = &TableSchema{Name: tableName}
			tables[k] = t
			order = append(order, k)
		}

		c := ColumnSchema{
			Name:      colName,
			DataType:  dataType,
			Nullable:  strings.EqualFold(isNullable, "YES"),
			IsPrimary: strings.EqualFold(colKey.String, "PRI"),
		}
		if colDefault.Valid {
			c.Default = colDefault.String
		}
		if extra.Valid {
			c.Extra = extra.String
		}
		t.Columns = append(t.Columns, c)
	}
	if err := rows.Err(); err != nil {
		return nil, failure.NewInternalError(err)
	}

	// Foreign keys
	const fkSQL = `
SELECT TABLE_NAME, COLUMN_NAME, CONSTRAINT_NAME, REFERENCED_TABLE_NAME, REFERENCED_COLUMN_NAME
FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
WHERE TABLE_SCHEMA = ? AND REFERENCED_TABLE_NAME IS NOT NULL
ORDER BY TABLE_NAME, COLUMN_NAME
`
	fkRows, err := db.QueryContext(ctx, fkSQL, dbName)
	if err != nil {
		return nil, failure.NewInternalError(err)
	}
	defer fkRows.Close()

	for fkRows.Next() {
		var tableName, colName, constraintName, refTable, refCol string
		if err := fkRows.Scan(&tableName, &colName, &constraintName, &refTable, &refCol); err != nil {
			return nil, failure.NewInternalError(err)
		}

		k := tableKey{schema: "", name: tableName}
		t, ok := tables[k]
		if !ok {
			t = &TableSchema{Name: tableName}
			tables[k] = t
			order = append(order, k)
		}
		t.ForeignKeys = append(t.ForeignKeys, ForeignKeySchema{
			ConstraintName: constraintName,
			Column:         colName,
			RefTable:       refTable,
			RefColumn:      refCol,
		})
	}
	if err := fkRows.Err(); err != nil {
		return nil, failure.NewInternalError(err)
	}

	out := &DatabaseSchema{Database: dbName, Tables: make([]TableSchema, 0, len(order))}
	for _, k := range order {
		if t := tables[k]; t != nil {
			out.Tables = append(out.Tables, *t)
		}
	}
	return out, nil
}
