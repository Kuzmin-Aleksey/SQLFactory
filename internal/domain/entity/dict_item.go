package entity

type DictItem struct {
	Id      int    `json:"id" db:"id"`
	DB      string `json:"db" db:"db"`
	Word    string `json:"word" db:"word"`
	Meaning string `json:"meaning" db:"meaning"`
}
