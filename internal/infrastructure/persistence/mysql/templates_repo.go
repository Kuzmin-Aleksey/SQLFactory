package mysql

import (
	"SQLFactory/internal/domain/entity"
	"SQLFactory/pkg/failure"
	"context"
	"database/sql"
	"errors"
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
	res, err := r.db.NamedExecContext(ctx, "INSERT INTO templates (db, title, query, table_data, chart_type) VALUES (:db, :title, :query, :table_data, :chart_type)", template)
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

func (r *TemplatesRepo) GetById(ctx context.Context, id int) (*entity.Template, error) {
	template := new(entity.Template)
	if err := r.db.GetContext(ctx, template, "SELECT * FROM templates WHERE id=?", id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, failure.NewNotFoundError(err)
		}
		return nil, failure.NewInternalError(err)
	}
	return template, nil
}

func (r *TemplatesRepo) UpdateTemplate(ctx context.Context, template *entity.Template) error {
	if _, err := r.db.NamedExecContext(ctx, "UPDATE templates SET title=:title, query=:query, chart_type=:chart_type WHERE id=:id", template); err != nil {
		return failure.NewInternalError(err)
	}
	return nil
}

func (r *TemplatesRepo) UpdateTemplateData(ctx context.Context, template *entity.Template) error {
	if _, err := r.db.NamedExecContext(ctx, "UPDATE templates SET table_data=:table_data WHERE id=:id", template); err != nil {
		return failure.NewInternalError(err)
	}
	return nil
}

func (r *TemplatesRepo) GetDBTemplates(ctx context.Context, db string) ([]entity.Template, error) {
	templates := []entity.Template{}
	if err := r.db.SelectContext(ctx, &templates, "SELECT * FROM templates WHERE db=?", db); err != nil {
		return nil, failure.NewInternalError(err)
	}
	return templates, nil
}

func (r *TemplatesRepo) DeleteTemplate(ctx context.Context, id int) error {
	if _, err := r.db.ExecContext(ctx, "DELETE FROM templates WHERE id=?", id); err != nil {
		return failure.NewInternalError(err)
	}
	return nil
}
