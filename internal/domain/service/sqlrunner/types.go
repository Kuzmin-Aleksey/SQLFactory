package sqlrunner

type QueryResult struct {
	Header []string   `json:"columns"`
	Data   [][]string `json:"rows"`
}

type QueryRequest struct {
	DBType   string `json:"db_type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	SQL      string `json:"sql"`
}
