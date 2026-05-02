package model

type QueryHistory struct {
	ID         int64  `json:"id"`
	Query      string `json:"query"`
	ExecutedAt string `json:"executedAt"`
}
