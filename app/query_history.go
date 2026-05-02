package app

import (
	"database/sql"
	"dbmx/model"
)

type QueryHistory struct {
	DB *sql.DB
}

func NewQueryHistory(db *sql.DB) *QueryHistory {
	return &QueryHistory{DB: db}
}

func (qh *QueryHistory) AddQueryHistory(query string) error {
	stmt := `INSERT INTO query_history (query) VALUES (?)`
	_, err := qh.DB.Exec(stmt, query)
	return err
}

func (qh *QueryHistory) GetQueryHistory() ([]model.QueryHistory, error) {
	query := `SELECT id, query, executed_at FROM query_history ORDER BY executed_at DESC LIMIT 50`
	rows, err := qh.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []model.QueryHistory
	for rows.Next() {
		var h model.QueryHistory
		err := rows.Scan(&h.ID, &h.Query, &h.ExecutedAt)
		if err != nil {
			return nil, err
		}
		history = append(history, h)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return history, nil
}
