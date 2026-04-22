package entity

import "time"

type HistoryItem struct {
	Id        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	DB        string    `json:"db" db:"db"`
	CreateAt  time.Time `json:"create_at" db:"create_at"`
	Title     string    `json:"title" db:"title"`
	Prompt    string    `json:"prompt,omitempty" db:"prompt"`
	Query     string    `json:"query,omitempty" db:"query"`
	Data      string    `json:"data,omitempty" db:"data"`
	Reasoning string    `json:"reasoning,omitempty" db:"reasoning"`
}
