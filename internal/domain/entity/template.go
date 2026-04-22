package entity

type Template struct {
	Id    int    `json:"id" db:"id"`
	DB    string `json:"db" db:"db"`
	Title string `json:"title" db:"title"`
	Query string `json:"query" db:"query"`
}
