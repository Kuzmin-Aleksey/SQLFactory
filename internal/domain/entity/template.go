package entity

import "SQLFactory/internal/domain/value"

type Template struct {
	Id        int             `json:"id" db:"id"`
	DB        string          `json:"db" db:"db"`
	Title     string          `json:"title" db:"title"`
	Query     string          `json:"query" db:"query"`
	TableData value.JsonValue `json:"table_data" db:"table_data"`
	ChartType string          `json:"chart_type" db:"chart_type"`
}
