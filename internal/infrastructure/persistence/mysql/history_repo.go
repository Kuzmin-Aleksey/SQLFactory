package mysql

import (
	"SQLFactory/internal/domain/entity"
	"SQLFactory/internal/domain/value"
	"SQLFactory/pkg/failure"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type HistoryRepo struct {
	db *sqlx.DB
}

func NewHistoryRepo(db *sqlx.DB) *HistoryRepo {
	return &HistoryRepo{
		db: db,
	}
}

func (r *HistoryRepo) SaveItem(ctx context.Context, item *entity.HistoryItem) error {
	res, err := r.db.NamedExecContext(ctx, "INSERT INTO history (user_id, db, create_at, title, prompt, query, table_data, chart_type, reasoning) VALUES (:user_id, :db, :create_at, :title, :prompt, :query, :table_data, :chart_type, :reasoning)", item)
	if err != nil {
		return failure.NewInternalError(err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return failure.NewInternalError(err)
	}

	item.Id = int(id)
	return nil
}

func (r *HistoryRepo) GetByDB(ctx context.Context, db string) ([]entity.HistoryItem, error) {
	items := []entity.HistoryItem{}
	if err := r.db.SelectContext(ctx, &items, "SELECT id, user_id, db, create_at, title FROM history WHERE db=?", db); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, failure.NewInternalError(err)
		}
	}
	return items, nil
}

func (r *HistoryRepo) GetItem(ctx context.Context, id int) (*entity.HistoryItem, error) {
	item := new(entity.HistoryItem)
	if err := r.db.GetContext(ctx, item, "SELECT * FROM history WHERE id=?", id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, failure.NewNotFoundError(fmt.Errorf("history %d not found", id))
		}
		return nil, failure.NewInternalError(err)
	}
	return item, nil
}

func (r *HistoryRepo) UpdateTableData(ctx context.Context, id int, data value.JsonValue) error {
	if _, err := r.db.ExecContext(ctx, "UPDATE history SET table_data=? WHERE id=?", data, id); err != nil {
		return failure.NewInternalError(err)
	}
	return nil
}

func (r *HistoryRepo) Delete(ctx context.Context, id int) error {
	if _, err := r.db.ExecContext(ctx, "DELETE FROM history WHERE id=?", id); err != nil {
		return failure.NewInternalError(err)
	}
	return nil
}
