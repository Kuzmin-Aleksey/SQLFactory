package mysql

import (
	"SQLFactory/internal/domain/entity"
	"SQLFactory/pkg/failure"
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
)

type DictRepo struct {
	db *sqlx.DB
}

func NewDictRepo(db *sqlx.DB) *DictRepo {
	return &DictRepo{
		db: db,
	}
}

func (r *DictRepo) Add(ctx context.Context, dictItem *entity.DictItem) error {
	res, err := r.db.NamedExecContext(ctx, "INSERT INTO dict (db, word, meaning) VALUES (:db, :word, :meaning)", dictItem)
	if err != nil {
		return failure.NewInternalError(err)
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return failure.NewInternalError(err)
	}
	dictItem.Id = int(lastID)

	return nil
}

func (r *DictRepo) Update(ctx context.Context, item *entity.DictItem) error {
	if _, err := r.db.NamedExecContext(ctx, "UPDATE dict SET word=:word, meaning=:meaning WHERE id=:id", item); err != nil {
		return failure.NewInternalError(err)
	}
	return nil
}

func (r *DictRepo) GetByDB(ctx context.Context, dbId string) (map[string]string, error) {
	dict := make(map[string]string)
	rows, err := r.db.QueryxContext(ctx, "SELECT word, meaning FROM dict WHERE db=?", dbId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dict, nil
		}
		return nil, failure.NewInternalError(err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			word    string
			meaning string
		)
		if err := rows.Scan(&word, &meaning); err != nil {
			return nil, failure.NewInternalError(err)
		}
		dict[word] = meaning
	}
	return dict, nil
}

func (r *DictRepo) Delete(ctx context.Context, id int) error {
	if _, err := r.db.NamedExecContext(ctx, "DELETE FROM dict WHERE id=:id", id); err != nil {
		return failure.NewInternalError(err)
	}
	return nil
}
