package mysql

import (
	"SQLFactory/internal/domain/entity"
	"SQLFactory/pkg/failure"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type LLMContextRepo struct {
	db *sqlx.DB
}

func NewLLMContextRepo(db *sqlx.DB) *LLMContextRepo {
	return &LLMContextRepo{
		db: db,
	}
}

func (r *LLMContextRepo) Save(ctx context.Context, llmContext *entity.LLMContext) error {
	res, err := r.db.NamedExecContext(ctx, "INSERT INTO llm_contexts (history_id, previous_id, role, content) VALUES (:history_id, :previous_id, :role, :content)", llmContext)
	if err != nil {
		return failure.NewInternalError(err)
	}
	lastID, _ := res.LastInsertId()
	llmContext.Id = int(lastID)
	return nil
}

func (r *LLMContextRepo) GetFullContextByHistoryId(ctx context.Context, historyId int) ([]entity.LLMContext, error) {
	var items []entity.LLMContext
	if err := r.db.SelectContext(ctx, &items, `
WITH RECURSIVE path AS (
    SELECT *
    FROM llm_contexts
    WHERE history_id = ?
    
    UNION ALL
    
    SELECT llm_contexts.*
    FROM llm_contexts
    INNER JOIN path p ON llm_contexts.id = p.previous_id
)
SELECT * FROM path ORDER BY id;
`, historyId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, failure.NewNotFoundError(fmt.Errorf("context with history_id=%d not found", historyId))
		}
		return nil, failure.NewInternalError(err)
	}
	return items, nil
}
