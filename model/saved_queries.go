package model

type SavedQuery struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Query   string `json:"query"`
	SavedAt string `json:"savedAt"`
}
