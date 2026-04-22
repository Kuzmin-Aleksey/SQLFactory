package entity

import (
	"SQLFactory/internal/domain/value"
	"time"
)

type HistoryItem struct {
	Id        int             `json:"id" db:"id"`
	UserID    int             `json:"user_id" db:"user_id"`
	DB        string          `json:"db" db:"db"`
	CreateAt  time.Time       `json:"create_at" db:"create_at"`
	Title     string          `json:"title" db:"title"`
	Prompt    string          `json:"prompt,omitempty" db:"prompt"`
	Query     string          `json:"query,omitempty" db:"query"`
	TableData value.JsonValue `json:"table_data,omitempty" db:"table_data"`
	ChartType string          `json:"chart_type,omitempty" db:"chart_type"`
	Reasoning value.JsonValue `json:"reasoning,omitempty" db:"reasoning"`
}
