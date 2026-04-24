package entity

type Role string

const (
	LLMRoleModel Role = "model"
	LLMRoleUser  Role = "user"
)

type LLMContext struct {
	Id         int    `json:"id" db:"id"`
	HistoryId  int    `json:"history_id" db:"history_id"`
	PreviousId *int   `json:"previous_id" db:"previous_id"`
	Role       Role   `json:"role" db:"role"`
	Content    string `json:"content" db:"content"`
}
