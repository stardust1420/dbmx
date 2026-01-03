package model

// Connection represents a row in the connections table in sqlite3 which represents a connection to a database engine server
type Connection struct {
	ID          int64
	Engine      string
	Host        string
	Port        string
	Username    string
	Password    string
	Database    string
	Name        string
	Env         string
	Color       string
	IsAdvanced  bool
	SSLMode     string
	ClientKey   []byte
	ClientCert  []byte
	RootCACert  []byte
	OverSSH     bool
	SSHHost     string
	SSHPort     string
	SSHUsername string
	SSHPassword string
	UseSSHKey   bool
	SSHKey      []byte

	// Only set for active connection

	IsActive bool
}

type ConnectionTable struct {
	ID       int64
	Name     string
	Env      string
	Engine   string
	Host     string
	Database string
}

type Database struct {
	// ID uniquely identifies the active database within a connection
	ID string
	// PostgresConnectionID is the primary key id of the postgres connection in the sqlite3 database
	ConnectionID   int64
	ConnectionName string
	Name           string
	Color          string

	// Only set for active connection

	// Active pool id
	PoolID   string
	IsActive bool

	// Tables and columns are set for the active database
	Tables  []string
	Columns []string
}

type Cell struct {
	Column string `json:"column"`
	Value  string `json:"value"`
}

type QueryResult struct {
	OK           bool     `json:"ok"`
	Columns      []string `json:"columns"`
	Rows         [][]Cell `json:"rows"`
	RowsAffected int64    `json:"rowsAffected"`
	Message      string   `json:"message"`
}

type Output struct {
	Columns []string `json:"columns"`
	Rows    [][]Cell `json:"rows"`
}

type Structure struct {
	Columns []string `json:"columns"`
	Rows    [][]Cell `json:"rows"`
}

type Indexes struct {
	Columns []string `json:"columns"`
	Rows    [][]Cell `json:"rows"`
}

type Rules struct {
	Columns []string `json:"columns"`
	Rows    [][]Cell `json:"rows"`
}

type TableInfo struct {
	Structure Structure `json:"structure"`
	Indexes   Indexes   `json:"indexes"`
	Rules     Rules     `json:"rules"`
}
