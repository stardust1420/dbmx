package app

import (
	"database/sql"
	"dbmx/model"
)

type SavedQueries struct {
	DB *sql.DB
}

func NewSavedQueries(db *sql.DB) *SavedQueries {
	return &SavedQueries{DB: db}
}

func (sq *SavedQueries) SaveQuery(title string, query string) error {
	stmt := `INSERT INTO saved_queries (title, query) VALUES (?, ?)`
	_, err := sq.DB.Exec(stmt, title, query)
	return err
}

func (sq *SavedQueries) GetSavedQueries() ([]model.SavedQuery, error) {
	query := `SELECT id, title, query, saved_at FROM saved_queries ORDER BY saved_at DESC LIMIT 50`
	rows, err := sq.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var queries []model.SavedQuery
	for rows.Next() {
		var q model.SavedQuery
		err := rows.Scan(&q.ID, &q.Title, &q.Query, &q.SavedAt)
		if err != nil {
			return nil, err
		}
		queries = append(queries, q)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return queries, nil
}

func (sq *SavedQueries) DeleteSavedQuery(id int64) error {
	stmt := `DELETE FROM saved_queries WHERE id = ?`
	_, err := sq.DB.Exec(stmt, id)
	return err
}
