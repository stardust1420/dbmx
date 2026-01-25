package app

import (
	"context"
	"database/sql"
	"dbmx/model"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Tabs struct {
	DB *sql.DB
	PM *PoolManager
}

func NewTabs(db *sql.DB, pm *PoolManager) *Tabs {
	return &Tabs{
		DB: db,
		PM: pm,
	}
}

func (t *Tabs) AddTab(activeDBID, activeDB, activeDBColor, tableName, tabType string, connID int64, dbName, connName string) (*model.Tab, error) {
	var active_db_id, active_db, active_db_color *string

	// Required for tab type table
	var conn_id *int64
	var db_name *string

	var tableColumnsString string
	var tableColumns []string

	name := "Editor"

	if tabType == "table" {
		if connID == 0 {
			return nil, errors.New("connection id is required for tab type table")
		}
		if dbName == "" {
			return nil, errors.New("database name is required for tab type table")
		}
		if tableName == "" {
			return nil, errors.New("table name is required for tab type table")
		}
		if activeDBID == "" {
			return nil, errors.New("active db pool id is required for tab type table")
		}

		conn_id = &connID
		db_name = &dbName
		name = tableName

		// Get the columns of the table

		// Convert active db id to uuid
		activePoolID, err := uuid.Parse(activeDBID)
		if err != nil {
			return nil, err
		}
		pool, exists := t.PM.GetPool(activePoolID)
		if !exists {
			return nil, errors.New("pool doesn't exist")
		}

		// Get all columns
		query := `
			SELECT column_name
			FROM information_schema.columns
			WHERE table_name = $1
			  AND table_schema = 'public';
		`
		rows, err := pool.Query(context.Background(), query, tableName)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		// Iterate through the rows
		for rows.Next() {
			var column string
			err := rows.Scan(&column)
			if err != nil {
				return nil, err
			}
			tableColumns = append(tableColumns, column)
		}

		// marshal the table columns into json
		tableColumnsJSON, err := json.Marshal(tableColumns)
		if err != nil {
			return nil, err
		}

		tableColumnsString = string(tableColumnsJSON)
	}

	if activeDBID != "" {
		active_db_id = &activeDBID
	}

	if activeDB != "" {
		active_db = &activeDB
	}
	if activeDBColor != "" {
		active_db_color = &activeDBColor
	}

	if !model.IsValidTabType(tabType) {
		return nil, errors.New("invalid tab type. Only editor and table are allowed.")
	}

	// Insert a new active tab
	query := `INSERT INTO tabs (name, editor, output, is_active, active_db_id, active_db, active_db_color, type, connection_id, db_name, connection_name, "select", "limit", "offset", "where", "order_by", "group_by", table_columns) VALUES (?, '', '', true, ?, ?, ?, ?, ?, ?, ?, '', '', '', '', '', '', ?);`
	result, err := t.DB.Exec(query, name, active_db_id, active_db, active_db_color, tabType, conn_id, db_name, connName, tableColumnsString)
	if err != nil {
		return nil, err
	}
	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Write an update query to set is_active to false for all other tabs
	updateQuery := `UPDATE tabs SET is_active = false WHERE id != ?`
	_, err = t.DB.Exec(updateQuery, insertedID)
	if err != nil {
		return nil, err
	}

	return &model.Tab{
		ID:               insertedID,
		Name:             name,
		IsActive:         true,
		ActiveDBID:       active_db_id,
		ActiveDB:         active_db,
		ActiveDBColor:    active_db_color,
		Type:             tabType,
		ConnectionID:     conn_id,
		DBName:           db_name,
		ConnectionName:   connName,
		TableColumnsList: tableColumns,
	}, nil
}

func (t *Tabs) SetActiveTab(id int64) (*model.Tab, error) {
	// Write an update query to set is_active to false for all other tabs
	updateQuery := `UPDATE tabs SET is_active = false WHERE id != ?`
	_, err := t.DB.Exec(updateQuery, id)
	if err != nil {
		return nil, err
	}

	// Write an update query to set is_active to true for the given tab
	updateQuery = `UPDATE tabs SET is_active = true WHERE id = ? RETURNING *`

	var tab model.Tab
	err = t.DB.QueryRow(updateQuery, id).Scan(&tab.ID, &tab.Name, &tab.Editor, &tab.Output, &tab.IsActive, &tab.ActiveDBID, &tab.ActiveDB, &tab.ActiveDBColor, &tab.Type, &tab.ConnectionID, &tab.DBName, &tab.ConnectionName, &tab.Select, &tab.Limit, &tab.Offset, &tab.Where, &tab.OrderBy, &tab.GroupBy, &tab.TableColumns)
	if err != nil {
		return nil, err
	}

	if len(tab.TableColumns) > 0 {
		err = json.Unmarshal([]byte(tab.TableColumns), &tab.TableColumnsList)
		if err != nil {
			return nil, err
		}
	}

	if len(tab.Output) == 0 {
		return &tab, nil
	}

	return &tab, nil
}

func (t *Tabs) GetAllTabs() ([]model.Tab, error) {
	// Query for all tabs
	query := `SELECT id, name, editor, output, is_active, active_db_id, active_db, active_db_color, type, connection_id, db_name, connection_name, "select", "limit", "offset", "where", "order_by", "group_by", table_columns FROM tabs`
	rows, err := t.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tabs []model.Tab
	for rows.Next() {
		var tab model.Tab
		err := rows.Scan(&tab.ID, &tab.Name, &tab.Editor, &tab.Output, &tab.IsActive, &tab.ActiveDBID, &tab.ActiveDB, &tab.ActiveDBColor, &tab.Type, &tab.ConnectionID, &tab.DBName, &tab.ConnectionName, &tab.Select, &tab.Limit, &tab.Offset, &tab.Where, &tab.OrderBy, &tab.GroupBy, &tab.TableColumns)
		if err != nil {
			return nil, err
		}

		if tab.IsActive {
			if len(tab.TableColumns) > 0 {
				err = json.Unmarshal([]byte(tab.TableColumns), &tab.TableColumnsList)
				if err != nil {
					return nil, err
				}
			}

			if len(tab.Output) > 0 {
				// Don't load output from DB, frontend will handle it
			}
		}

		tabs = append(tabs, tab)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tabs, nil
}

func (t *Tabs) DeleteTab(id int64) (*model.Tab, error) {
	// Check if the tab is active
	query := `SELECT is_active FROM tabs WHERE id = ?`
	var isActive bool
	err := t.DB.QueryRow(query, id).Scan(&isActive)
	if err != nil {
		return nil, err
	}

	// Tab to be returned
	var tab model.Tab

	var isLastTab bool

	// If the tab is active, set the first tab as active where id != id and return that tab
	if isActive {
		// Check for the count of tabs to find if it's the last tab
		query = `SELECT COUNT(*) FROM tabs`
		var count int64
		err := t.DB.QueryRow(query).Scan(&count)
		if err != nil {
			return nil, err
		}

		// If it's not the last tab, set the first tab as active
		if count > 1 {
			query = `UPDATE tabs SET is_active = true WHERE id = (SELECT id FROM tabs WHERE id != ? LIMIT 1) RETURNING *`
			row := t.DB.QueryRow(query, id)

			err = row.Scan(&tab.ID, &tab.Name, &tab.Editor, &tab.Output, &tab.IsActive, &tab.ActiveDBID, &tab.ActiveDB, &tab.ActiveDBColor, &tab.Type, &tab.ConnectionID, &tab.DBName, &tab.ConnectionName, &tab.Select, &tab.Limit, &tab.Offset, &tab.Where, &tab.OrderBy, &tab.GroupBy, &tab.TableColumns)
			if err != nil {
				return nil, err
			}

			if len(tab.TableColumns) > 0 {
				err = json.Unmarshal([]byte(tab.TableColumns), &tab.TableColumnsList)
				if err != nil {
					return nil, err
				}
			}

			if len(tab.Output) > 0 {
				// Don't load output from DB
			}
		} else {
			isLastTab = true
		}
	}

	// Delete the tab
	query = `DELETE FROM tabs WHERE id = ?`
	_, err = t.DB.Exec(query, id)
	if err != nil {
		return nil, err
	}

	if isActive && !isLastTab {
		return &tab, nil
	}

	return nil, nil
}

func (t *Tabs) UpdateTabEditorContent(id int64, editor string, selectQuery, limit, offset, where, orderBy, groupBy string) error {
	query := `UPDATE tabs SET editor = ?, "select" = ?, "limit" = ?, "offset" = ?, "where" = ?, "order_by" = ?, "group_by" = ? WHERE id = ?`
	_, err := t.DB.Exec(query, editor, selectQuery, limit, offset, where, orderBy, groupBy, id)
	if err != nil {
		return err
	}

	return nil
}

func (t *Tabs) SaveActiveDBProps(id int64, activeDBID, activeDB, activeDBColor string) error {
	var active_db_id, active_db, active_db_color *string
	if activeDBID != "" {
		active_db_id = &activeDBID
	}
	if activeDB != "" {
		active_db = &activeDB
	}
	if activeDBColor != "" {
		active_db_color = &activeDBColor
	}

	// NOTE: Only update for tab type editor
	// Update the active db properties in the tab of type editor
	query := `UPDATE tabs SET active_db_id = ?, active_db = ?, active_db_color = ? WHERE id = ? AND type = 'editor'`
	_, err := t.DB.Exec(query, active_db_id, active_db, active_db_color, id)
	if err != nil {
		return err
	}

	return nil
}
