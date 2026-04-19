package model

type Tab struct {
	ID       int64
	Name     string
	Editor   string
	Output   string
	IsActive bool

	// This is the active pool id
	ActiveDBID *string

	// This is the active server connection + database name
	ActiveDB *string

	// Color of the active server
	ActiveDBColor *string
	Type          string

	// Output
	Columns     []string `json:"columns"`
	Rows        [][]Cell `json:"rows"`
	TotalRows   int64    `json:"totalRows"`
	CurrentPage int64    `json:"currentPage"`

	// Required if type is table
	ConnectionID   *int64
	ConnectionName string
	DBName         *string
	Select         string
	Limit          string
	Offset         string
	Where          string
	OrderBy        string
	GroupBy        string
	TableColumns   string

	// To be passed to frontend
	TableColumnsList []string

	// AI chat
	AIChat []AIMsg
}

type AIMsg struct {
	ID        string
	Role      string
	Content   string
	CreatedAt string
}

var validTypes = map[string]struct{}{
	"editor": {},
	"table":  {},
}

func IsValidTabType(t string) bool {
	_, ok := validTypes[t]
	return ok
}
