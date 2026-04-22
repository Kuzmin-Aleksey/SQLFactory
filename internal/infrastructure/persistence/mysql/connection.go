package mysql

import (
	"SQLFactory/internal/config"
	"context"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"time"
)

func Connect(cfg config.MySqlConfig) (*sqlx.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ConnectTimeout)*time.Second)
	defer cancel()

	dataSource := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&multiStatements=true", cfg.User, cfg.Password, cfg.Addr, cfg.Schema)

	db, err := sqlx.ConnectContext(ctx, "mysql", dataSource)
	if err != nil {
		return nil, err
	}

	return db, nil
}
