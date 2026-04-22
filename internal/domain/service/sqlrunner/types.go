package sqlrunner

type QueryResult struct {
	Header []string   `json:"columns"`
	Data   [][]string `json:"rows"`
}

type ConnectionRequest struct {
	DBType   string `json:"db_type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type QueryRequest struct {
	SQL string `json:"sql"`
	ConnectionRequest
}

type DatabaseSchema struct {
	Database string        `json:"database"`
	Tables   []TableSchema `json:"tables"`
}

type TableSchema struct {
	Schema      string             `json:"schema,omitempty"`
	Name        string             `json:"name"`
	Columns     []ColumnSchema     `json:"columns"`
	ForeignKeys []ForeignKeySchema `json:"foreign_keys,omitempty"`
}

type ColumnSchema struct {
	Name      string `json:"name"`
	DataType  string `json:"data_type"`
	Nullable  bool   `json:"nullable"`
	Default   string `json:"default,omitempty"`
	Extra     string `json:"extra,omitempty"`
	IsPrimary bool   `json:"is_primary,omitempty"`
}

type ForeignKeySchema struct {
	ConstraintName string `json:"constraint_name,omitempty"`
	Column         string `json:"column"`
	RefSchema      string `json:"ref_schema,omitempty"`
	RefTable       string `json:"ref_table"`
	RefColumn      string `json:"ref_column"`
}
