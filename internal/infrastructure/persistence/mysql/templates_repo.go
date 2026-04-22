package mysql

import (
	"SQLFactory/internal/domain/entity"
	"SQLFactory/pkg/failure"
	"context"
	"github.com/jmoiron/sqlx"
)

type TemplatesRepo struct {
	db *sqlx.DB
}

func NewTemplatesRepo(db *sqlx.DB) *TemplatesRepo {
	return &TemplatesRepo{
		db: db,
	}
}

func (r *TemplatesRepo) SaveTemplate(ctx context.Context, template *entity.Template) error {
	res, err := r.db.NamedExecContext(ctx, "INSERT INTO templates (db, title, query) VALUES (:db_id, :title, :query)", template)
	if err != nil {
		return failure.NewInternalError(err)
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return failure.NewInternalError(err)
	}
	template.Id = int(lastID)

	return nil
}

func (r *TemplatesRepo) UpdateTemplate(ctx context.Context, template *entity.Template) error {
	if _, err := r.db.NamedExecContext(ctx, "UPDATE templates SET title=:title, query=:query WHERE id=:id", template); err != nil {
		return failure.NewInternalError(err)
	}
	return nil
}

func (r *TemplatesRepo) GetDBTemplates(ctx context.Context, db string) ([]entity.Template, error) {
	var templates []entity.Template
	if err := r.db.SelectContext(ctx, &templates, "SELECT * FROM templates WHERE db=?", db); err != nil {
		return nil, failure.NewInternalError(err)
	}
	return templates, nil
}

func (r *TemplatesRepo) DeleteTemplate(ctx context.Context, id int) error {
	if _, err := r.db.NamedExecContext(ctx, "DELETE FROM templates WHERE id=:id", id); err != nil {
		return failure.NewInternalError(err)
	}
	return nil
}
