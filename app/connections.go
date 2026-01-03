package app

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"dbmx/model"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"

	"github.com/jackc/pgx/v5"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

type Connections struct {
	DB *sql.DB
	PM *PoolManager
}

func NewConnections(db *sql.DB, pm *PoolManager) *Connections {
	return &Connections{
		DB: db,
		PM: pm,
	}
}

func (c *Connections) TypeConnectionTable() *model.ConnectionTable {
	return &model.ConnectionTable{}
}

func (m *Connections) GetSqlite3Version() string {
	// Get the version of SQLite
	var sqliteVersion string
	err := m.DB.QueryRow("SELECT sqlite_version()").Scan(&sqliteVersion)
	if err != nil {
		log.Fatal(err)
	}

	return sqliteVersion
}

func (m *Connections) GetAllConnections() ([]model.Connection, error) {
	// Get all connections
	connections := []model.Connection{}

	rows, err := m.DB.Query("SELECT * FROM connections")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var connection model.Connection
		err := rows.Scan(
			&connection.ID,
			&connection.Engine,
			&connection.Host,
			&connection.Port,
			&connection.Username,
			&connection.Password,
			&connection.Database,
			&connection.Name,
			&connection.Env,
			&connection.Color,
			&connection.IsAdvanced,
			&connection.SSLMode,
			&connection.ClientKey,
			&connection.ClientCert,
			&connection.RootCACert,
			&connection.OverSSH,
			&connection.SSHHost,
			&connection.SSHPort,
			&connection.SSHUsername,
			&connection.SSHPassword,
			&connection.UseSSHKey,
			&connection.SSHKey,
		)
		if err != nil {
			return nil, errors.Wrap(err, "unable to read resultant rows into connection variable")
		}
		connections = append(connections, connection)
	}
	if row_err := rows.Err(); row_err != nil {
		return nil, errors.Wrap(row_err, "unable to read rows")
	}

	return connections, nil
}

func BuildPostgresConnConfig(c model.Connection) (*pgx.ConnConfig, error) {
	if c.Database == "" {
		c.Database = "postgres"
	}

	// Build connection string using the credentials
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
	)

	sslMode := strings.ToLower(strings.TrimSpace(c.SSLMode))

	if sslMode != "" {
		connString += fmt.Sprintf("?sslmode=%s", sslMode)
	}

	config, err := pgx.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	if c.IsAdvanced {
		tlsConfig := &tls.Config{}

		// Root CA
		if len(c.RootCACert) > 0 {
			rootPool := x509.NewCertPool()
			if !rootPool.AppendCertsFromPEM(c.RootCACert) {
				return nil, fmt.Errorf("failed to parse root CA cert")
			}
			tlsConfig.RootCAs = rootPool
		}

		// Client cert + key (mutual TLS)
		if len(c.ClientKey) > 0 && len(c.ClientCert) > 0 {
			cert, err := tls.X509KeyPair(c.ClientCert, c.ClientKey)
			if err != nil {
				return nil, err
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}

		config.TLSConfig = tlsConfig
	}

	if c.OverSSH {
		var sshClient *ssh.Client

		var authMethods []ssh.AuthMethod

		if c.UseSSHKey {
			signer, err := ssh.ParsePrivateKey(c.SSHKey)
			if err != nil {
				return nil, err
			}
			authMethods = append(authMethods, ssh.PublicKeys(signer))
		} else {
			authMethods = append(authMethods, ssh.Password(c.SSHPassword))
		}

		sshConfig := &ssh.ClientConfig{
			User:            c.SSHUsername,
			Auth:            authMethods,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(), // You may want strict checking later
			Timeout:         10 * time.Second,
		}

		sshClient, err = ssh.Dial(
			"tcp",
			net.JoinHostPort(c.SSHHost, c.SSHPort),
			sshConfig,
		)
		if err != nil {
			return nil, err
		}

		config.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return sshClient.Dial(network, addr)
		}
	}

	return config, nil
}

func (m *Connections) TestConnectPostgres(c model.Connection) (bool, error) {
	config, err := BuildPostgresConnConfig(c)
	if err != nil {
		return false, err
	}

	// Establish a connection
	ctx := context.Background()

	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return false, err
	}
	defer conn.Close(ctx)

	// Execute a simple query
	var greeting string
	err = conn.QueryRow(ctx, "SELECT 'Connection Successful!' AS success").Scan(&greeting)
	if err != nil {
		return false, errors.Wrap(err, "failed to query database")
	}

	// Print the result
	fmt.Printf("Greeting: %s\n", greeting)

	return true, nil
}

func (m *Connections) AddPostgresConnection(c model.Connection) (bool, error) {
	if c.Database == "" {
		c.Database = "postgres"
	}

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM connections WHERE name = ? AND engine = ? AND env = ?)`

	// Execute the query
	err := m.DB.QueryRow(query, c.Name, c.Engine, c.Env).Scan(&exists)
	if err != nil {
		return false, err
	}

	if exists {
		return false, errors.New("Connection name already exists. Please choose a different name")
	}

	query = `
		INSERT INTO connections(
			engine,
			host,
			port,
			username,
			password,
			database,
			name,
			env,
			color,
			is_advanced,
			ssl_mode,
			client_key,
			client_cert,
			root_ca_cert,
			over_ssh,
			ssh_host,
			ssh_port,
			ssh_username,
			ssh_password,
			use_ssh_key,
			ssh_key
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	insertStatement, err := m.DB.Prepare(query)
	if err != nil {
		return false, errors.Wrap(err, "failed to prepare query to insert new connection in connections")
	}

	_, err = insertStatement.Exec(
		c.Engine,
		c.Host,
		c.Port,
		c.Username,
		c.Password,
		c.Database,
		c.Name,
		c.Env,
		c.Color,
		c.IsAdvanced,
		c.SSLMode,
		c.ClientKey,
		c.ClientCert,
		c.RootCACert,
		c.OverSSH,
		c.SSHHost,
		c.SSHPort,
		c.SSHUsername,
		c.SSHPassword,
		c.UseSSHKey,
		c.SSHKey,
	)
	if err != nil {
		return false, errors.Wrap(err, "failed to insert new connection in connections")
	}

	return true, nil
}

func (c *Connections) RefreshPostgresDatabase(id int64, dbID, dbName, poolID string) (*model.Database, error) {
	poolIDUUID, err := uuid.Parse(poolID)
	if err != nil {
		return nil, err
	}

	// Get all tables
	tables, err := c.GetAllPostgresTables(poolIDUUID)
	if err != nil {
		return nil, err
	}

	return &model.Database{
		ID:           dbID,
		Name:         dbName,
		ConnectionID: id,
		PoolID:       poolID,
		IsActive:     true,
		Tables:       tables,
	}, nil
}

// This func is used to connect a specific database within a server
// id is the postgres connection id primary key in the sqlite3 database
// dbID uniquely identifies the active database within a connection
func (c *Connections) EstablishPostgresDatabaseConnection(id int64, dbName string) (*model.Database, error) {
	var conn model.Connection

	err := c.DB.QueryRow("SELECT * FROM connections WHERE id = ?", id).Scan(
		&conn.ID,
		&conn.Engine,
		&conn.Host,
		&conn.Port,
		&conn.Username,
		&conn.Password,
		&conn.Database,
		&conn.Name,
		&conn.Env,
		&conn.Color,
		&conn.IsAdvanced,
		&conn.SSLMode,
		&conn.ClientKey,
		&conn.ClientCert,
		&conn.RootCACert,
		&conn.OverSSH,
		&conn.SSHHost,
		&conn.SSHPort,
		&conn.SSHUsername,
		&conn.SSHPassword,
		&conn.UseSSHKey,
		&conn.SSHKey,
	)
	if err != nil {
		return nil, err
	}

	cfg, err := BuildPostgresConnConfig(conn)
	if err != nil {
		return nil, err
	}

	activePoolID := uuid.New()

	// Establish connection and add pool to active pool manager
	_, err = c.PM.AddPool(activePoolID, cfg)
	if err != nil {
		return nil, err
	}

	// Get all table names for suggestions
	tables, err := c.GetAllPostgresTables(activePoolID)
	if err != nil {
		return nil, err
	}

	// Get all columns of all tables for suggestions
	columns, err := c.GetAllDatabaseColumns(activePoolID)
	if err != nil {
		return nil, err
	}

	activeDB := conn.Name + " - " + dbName

	// Save the active db properties in all the tabs with type editor if active db properties are null
	_, err = c.DB.Exec("UPDATE tabs SET active_db_id = ?, active_db = ?, active_db_color = ? WHERE active_db_id IS NULL AND type = 'editor'", activePoolID.String(), activeDB, conn.Color)
	if err != nil {
		return nil, err
	}

	// In case of table rows, find all the rows with type table where
	// active_db_id is null and postgres_connection_id and database matches
	// set the active pool id and active db properties in such tabs
	_, err = c.DB.Exec("UPDATE tabs SET active_db_id = ?, active_db = ?, active_db_color = ? WHERE active_db_id IS NULL AND type = 'table' AND connection_id = ? AND db_name = ?", activePoolID.String(), activeDB, conn.Color, id, dbName)
	if err != nil {
		return nil, err
	}

	return &model.Database{
		Name:           dbName,
		ConnectionID:   id,
		ConnectionName: conn.Name,
		Color:          conn.Color,
		PoolID:         activePoolID.String(),
		IsActive:       true,
		Tables:         tables,
		Columns:        columns,
	}, nil
}

// This func is used to connect to a server
func (c *Connections) EstablishPostgresConnection(id int64) ([]model.Database, error) {
	var conn model.Connection

	err := c.DB.QueryRow("SELECT * FROM connections WHERE id = ?", id).Scan(
		&conn.ID,
		&conn.Engine,
		&conn.Host,
		&conn.Port,
		&conn.Username,
		&conn.Password,
		&conn.Database,
		&conn.Name,
		&conn.Env,
		&conn.Color,
		&conn.IsAdvanced,
		&conn.SSLMode,
		&conn.ClientKey,
		&conn.ClientCert,
		&conn.RootCACert,
		&conn.OverSSH,
		&conn.SSHHost,
		&conn.SSHPort,
		&conn.SSHUsername,
		&conn.SSHPassword,
		&conn.UseSSHKey,
		&conn.SSHKey,
	)
	if err != nil {
		return nil, err
	}

	cfg, err := BuildPostgresConnConfig(conn)
	if err != nil {
		return nil, err
	}

	activePoolID := uuid.New()

	// Establish connection and add pool to active pool manager
	_, err = c.PM.AddPool(activePoolID, cfg)
	if err != nil {
		return nil, err
	}

	activeDB := conn.Name + " - " + conn.Database

	// Save the active db properties in all the tabs with type editor if active db properties are null
	_, err = c.DB.Exec("UPDATE tabs SET active_db_id = ?, active_db = ?, active_db_color = ? WHERE active_db_id IS NULL AND type = 'editor'", activePoolID.String(), activeDB, conn.Color)
	if err != nil {
		return nil, err
	}

	// In case of table rows, find all the rows with type table where
	// active_db_id is null and postgres_connection_id and database matches
	// set the active pool id and active db properties in such tabs
	_, err = c.DB.Exec("UPDATE tabs SET active_db_id = ?, active_db = ?, active_db_color = ? WHERE active_db_id IS NULL AND type = 'table' AND connection_id = ? AND db_name = ?", activePoolID.String(), activeDB, conn.Color, id, conn.Database)
	if err != nil {
		return nil, err
	}

	return c.GetPostgresServerDatabases(id, activePoolID, conn.Database, conn.Name, conn.Color)
}

// Here, pool means the active connection to the database server
func (c *Connections) GetPostgresServerDatabases(connectionID int64, activePoolID uuid.UUID, activeDatabase, connectionName, color string) ([]model.Database, error) {
	pool, exists := c.PM.GetPool(activePoolID)
	if !exists {
		return nil, errors.New("pool doesn't exist")
	}

	// Get Tables of active database
	tables, err := c.GetAllPostgresTables(activePoolID)
	if err != nil {
		return nil, err
	}

	// Get Columns of active database
	columns, err := c.GetAllDatabaseColumns(activePoolID)
	if err != nil {
		return nil, err
	}

	// Get all database names of the active connection
	rows, err := pool.Query(context.TODO(), "SELECT datname FROM pg_database")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Slice to hold results
	var databases []model.Database

	// Iterate through the rows
	for rows.Next() {
		var database model.Database
		err := rows.Scan(&database.Name)
		if err != nil {
			return nil, err
		}
		database.ID = "db_" + uuid.New().String()
		database.ConnectionID = connectionID
		database.ConnectionName = connectionName
		database.Color = color
		if database.Name == activeDatabase {
			database.PoolID = activePoolID.String()
			database.IsActive = true
			database.Tables = tables
			database.Columns = columns
		}

		databases = append(databases, database)
	}

	// Check for any error encountered during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return databases, nil
}

func (c *Connections) GetAllPostgresTables(activePoolID uuid.UUID) ([]string, error) {
	pool, exists := c.PM.GetPool(activePoolID)
	if !exists {
		return nil, errors.New("pool doesn't exist")
	}

	// Get all tables
	rows, err := pool.Query(context.TODO(), "SELECT tablename FROM pg_tables WHERE schemaname = 'public'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Slice to hold results
	var tables []string

	// Iterate through the rows
	for rows.Next() {
		var table string
		err := rows.Scan(&table)
		if err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	// Check for any error encountered during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}

// Get all columns of the active database across all tables
func (c *Connections) GetAllDatabaseColumns(activePoolID uuid.UUID) ([]string, error) {
	pool, exists := c.PM.GetPool(activePoolID)
	if !exists {
		return nil, errors.New("pool doesn't exist")
	}

	// Get all columns
	query := `
		SELECT DISTINCT
			column_name
		FROM
			information_schema.columns
		WHERE
			table_schema NOT IN ('information_schema', 'pg_catalog')
	`
	rows, err := pool.Query(context.TODO(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Slice to hold results
	var columns []string

	// Iterate through the rows
	for rows.Next() {
		var column string
		err := rows.Scan(&column)
		if err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}

	// Check for any error encountered during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return columns, nil
}

func (c *Connections) TerminatePostgresDatabaseConnection(activePoolID string) (bool, error) {
	activePoolIDUUID, err := uuid.Parse(activePoolID)
	if err != nil {
		return false, err
	}
	// Remove the db pool from active pools
	err = c.PM.DeletePool(activePoolIDUUID)
	if err != nil {
		return false, err
	}

	// Remove the pool from all the tabs in which it's saved
	_, err = c.DB.Exec("UPDATE tabs SET active_db_id = NULL, active_db = NULL, active_db_color = NULL WHERE active_db_id = ?", activePoolID)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (c *Connections) TerminateAllDatabaseConnections() error {
	c.PM.mu.Lock()
	defer c.PM.mu.Unlock()

	activeDBIds := []string{}

	for id, pool := range c.PM.Pools {
		activeDBIds = append(activeDBIds, id.String())
		pool.Close()
		delete(c.PM.Pools, id)
	}

	// Build placeholders (?, ?, ?)
	placeholders := strings.Repeat("?,", len(activeDBIds))
	placeholders = strings.TrimRight(placeholders, ",")

	// Remove the pool from all the tabs in which it's saved
	// Construct the query
	query := fmt.Sprintf(
		"UPDATE tabs SET active_db_id = NULL, active_db = NULL, active_db_color = NULL WHERE active_db_id IN (%s)",
		placeholders,
	)

	// Convert []string to []interface{}
	args := make([]interface{}, len(activeDBIds))
	for i, v := range activeDBIds {
		args[i] = v
	}

	// Execute the query with placeholders
	_, err := c.DB.Exec(query, args...)
	if err != nil {
		return err
	}

	return nil
}

// Function to check if a query is a write operation
func isWriteOperation(query string) bool {
	// List of SQL keywords for write operations in lowercase
	writeKeywords := []string{"insert", "update", "delete", "alter", "create", "drop", "truncate"}

	// Convert query to lowercase for case-insensitive comparison
	query = strings.ToLower(strings.TrimSpace(query))

	// Check if the query starts with any write keyword
	for _, keyword := range writeKeywords {
		if strings.HasPrefix(query, keyword) {
			return true
		}
	}
	return false
}

func (c *Connections) ExecuteQuery(activePoolID uuid.UUID, query string, tabID int64) *model.QueryResult {
	pool, exists := c.PM.GetPool(activePoolID)
	if !exists {
		return &model.QueryResult{OK: false, Message: "pool doesn't exist"}
	}

	ctx := context.Background()

	response := &model.QueryResult{OK: true}

	normalizedQuery := strings.ToLower(strings.TrimSpace(query))

	isWrite := isWriteOperation(normalizedQuery)

	if isWrite {
		// Use Exec for write operations
		tag, err := pool.Exec(ctx, query)
		if err != nil {
			return &model.QueryResult{
				OK:           true,
				Message:      err.Error(),
				RowsAffected: int64(0),
				Columns:      []string{"Error"},
				Rows:         [][]model.Cell{{model.Cell{Column: "Error", Value: err.Error()}}},
			}
		}
		response.RowsAffected = tag.RowsAffected()
		response.Columns = []string{"Rows Affected"}
		response.Rows = [][]model.Cell{{model.Cell{Column: "Rows Affected", Value: fmt.Sprintf("%d", response.RowsAffected)}}}
	} else {
		// Use Query for read operations
		resultRows, err := pool.Query(ctx, query)
		if err != nil {
			return &model.QueryResult{
				OK:           true,
				Message:      err.Error(),
				RowsAffected: int64(0),
				Columns:      []string{"Error"},
				Rows:         [][]model.Cell{{model.Cell{Column: "Error", Value: err.Error()}}},
			}
		}
		defer resultRows.Close()

		columns := resultRows.FieldDescriptions()
		columnNames := make([]string, len(columns))
		for i, column := range columns {
			columnNames[i] = string(column.Name)
		}
		response.Columns = columnNames

		var rows [][]model.Cell

		for resultRows.Next() {
			row, err := resultRows.Values()
			if err != nil {
				return &model.QueryResult{OK: false, Message: err.Error()}
			}

			cells := []model.Cell{}
			for i, cell := range row {
				newCell := model.Cell{
					Column: columnNames[i],
				}
				switch v := cell.(type) {
				case []byte:
					newCell.Value = string(v)
				case time.Time:
					newCell.Value = v.Format(time.RFC3339)
				case nil:
					newCell.Value = "NULL"
				case [16]uint8:
					newCell.Value = uuid.UUID(v).String()
				case string:
					if v == "" {
						newCell.Value = "EMPTY"
					}
					newCell.Value = v
				default:
					newCell.Value = fmt.Sprintf("%v", v)
				}
				cells = append(cells, newCell)
			}
			rows = append(rows, cells)
		}

		if err := resultRows.Err(); err != nil {
			return &model.QueryResult{OK: false, Message: err.Error()}
		}

		response.Rows = rows
	}

	output := &model.Output{
		Columns: response.Columns,
		Rows:    response.Rows,
	}

	go c.UpdateTabOutput(tabID, output)

	return response
}

func (c *Connections) GetTableData(activePoolID uuid.UUID, tabID int64, tableName, selectQuery, limit, offset, where, orderBy, groupBy string) *model.QueryResult {
	pool, exists := c.PM.GetPool(activePoolID)
	if !exists {
		return &model.QueryResult{OK: false, Message: "pool doesn't exist"}
	}

	ctx := context.Background()

	response := &model.QueryResult{OK: true}

	setLimit := strconv.Itoa(100)
	if strings.TrimSpace(limit) != "" {
		limitInt, err := strconv.Atoi(strings.TrimSpace(limit))
		if err != nil {
			return &model.QueryResult{OK: false, Message: "limit is not a number"}
		}
		if limitInt > 100 {
			return &model.QueryResult{OK: false, Message: "limit cannot be greater than 100"}
		}
		setLimit = strings.TrimSpace(limit)
	}

	if strings.TrimSpace(selectQuery) == "" {
		selectQuery = "*"
	}

	query := fmt.Sprintf("SELECT %s FROM %s", selectQuery, tableName)
	if strings.TrimSpace(where) != "" {
		query += fmt.Sprintf(" WHERE %s", strings.TrimSpace(where))
	}
	if strings.TrimSpace(groupBy) != "" {
		query += fmt.Sprintf(" GROUP BY %s", strings.TrimSpace(groupBy))
	}
	if strings.TrimSpace(orderBy) != "" {
		query += fmt.Sprintf(" ORDER BY %s", strings.TrimSpace(orderBy))
	}

	// Set limit
	query += fmt.Sprintf(" LIMIT %s", setLimit)

	if strings.TrimSpace(offset) != "" {
		query += fmt.Sprintf(" OFFSET %s", strings.TrimSpace(offset))
	}

	// Use Query for read operations
	resultRows, err := pool.Query(ctx, query)
	if err != nil {
		return &model.QueryResult{
			OK:           true,
			Message:      err.Error(),
			RowsAffected: int64(0),
			Columns:      []string{"Error"},
			Rows:         [][]model.Cell{{model.Cell{Column: "Error", Value: err.Error()}}},
		}
	}
	defer resultRows.Close()

	columns := resultRows.FieldDescriptions()
	columnNames := make([]string, len(columns))
	for i, column := range columns {
		columnNames[i] = string(column.Name)
	}
	response.Columns = columnNames

	var rows [][]model.Cell

	for resultRows.Next() {
		row, err := resultRows.Values()
		if err != nil {
			return &model.QueryResult{OK: false, Message: err.Error()}
		}

		cells := []model.Cell{}
		for i, cell := range row {
			newCell := model.Cell{
				Column: columnNames[i],
			}
			switch v := cell.(type) {
			case []byte:
				newCell.Value = string(v)
			case time.Time:
				newCell.Value = v.Format(time.RFC3339)
			case nil:
				newCell.Value = "NULL"
			case [16]uint8:
				newCell.Value = uuid.UUID(v).String()
			case string:
				if v == "" {
					newCell.Value = "EMPTY"
				}
				newCell.Value = v
			default:
				newCell.Value = fmt.Sprintf("%v", v)
			}
			cells = append(cells, newCell)
		}
		rows = append(rows, cells)
	}

	if err := resultRows.Err(); err != nil {
		return &model.QueryResult{OK: false, Message: err.Error()}
	}

	response.Rows = rows

	output := &model.Output{
		Columns: response.Columns,
		Rows:    response.Rows,
	}

	go c.UpdateTabOutput(tabID, output)

	return response
}

func (c *Connections) UpdateTabOutput(tabID int64, output *model.Output) {
	jsonOutput, err := json.Marshal(output)
	if err != nil {
		fmt.Println(err)
	}

	query := `UPDATE tabs SET output = ? WHERE id = ?`
	_, err = c.DB.Exec(query, string(jsonOutput), tabID)
	if err != nil {
		fmt.Println(err)
	}
}

func (c *Connections) GetTableInfo(activePoolID uuid.UUID, tableName string) (*model.TableInfo, error) {
	pool, exists := c.PM.GetPool(activePoolID)
	if !exists {
		return nil, errors.New("pool doesn't exist")
	}

	ctx := context.Background()

	// Get table structure
	query := `
		WITH fk_info AS (
			SELECT
				kcu.table_schema,
				kcu.table_name,
				kcu.column_name,
				string_agg(ccu.table_name || '.' || ccu.column_name, ', ') AS foreign_keys
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu
				ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
			JOIN information_schema.constraint_column_usage ccu
				ON ccu.constraint_name = tc.constraint_name
			AND ccu.table_schema = tc.table_schema
			WHERE tc.constraint_type = 'FOREIGN KEY'
			GROUP BY kcu.table_schema, kcu.table_name, kcu.column_name
		),
		check_info AS (
			SELECT
				tc.table_schema,
				tc.table_name,
				kcu.column_name,
				string_agg(cc.check_clause, ' AND ') AS check_clauses
			FROM information_schema.table_constraints tc
			JOIN information_schema.constraint_column_usage kcu
				ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
			JOIN information_schema.check_constraints cc
				ON tc.constraint_name = cc.constraint_name
			WHERE tc.constraint_type = 'CHECK'
			GROUP BY tc.table_schema, tc.table_name, kcu.column_name
		)
		SELECT 
			c.column_name,
			CASE 
				WHEN c.character_maximum_length IS NOT NULL 
					THEN c.data_type || '(' || c.character_maximum_length || ')'
				ELSE c.data_type
			END AS data_type,
			c.is_nullable,
			c.column_default,
			ch.check_clauses,
			fk.foreign_keys,
			pgd.description AS column_comment
		FROM information_schema.columns c
		LEFT JOIN check_info ch
			ON c.table_schema = ch.table_schema
		AND c.table_name = ch.table_name
		AND c.column_name = ch.column_name
		LEFT JOIN fk_info fk
			ON c.table_schema = fk.table_schema
		AND c.table_name = fk.table_name
		AND c.column_name = fk.column_name
		LEFT JOIN pg_catalog.pg_statio_all_tables st
			ON c.table_schema = st.schemaname
		AND c.table_name = st.relname
		LEFT JOIN pg_catalog.pg_description pgd
			ON pgd.objoid = st.relid
		AND pgd.objsubid = c.ordinal_position
		WHERE c.table_name = $1
		AND c.table_schema = 'public'
		ORDER BY c.ordinal_position;
	`
	resultRows, err := pool.Query(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer resultRows.Close()

	var structure model.Structure

	fieldDescriptions := resultRows.FieldDescriptions()
	columns := make([]string, len(fieldDescriptions))
	for i, column := range fieldDescriptions {
		columns[i] = string(column.Name)
	}
	structure.Columns = columns

	var rows [][]model.Cell

	for resultRows.Next() {
		row, err := resultRows.Values()
		if err != nil {
			return nil, err
		}

		cells := []model.Cell{}
		for i, cell := range row {
			newCell := model.Cell{
				Column: columns[i],
			}
			switch v := cell.(type) {
			case []byte:
				newCell.Value = string(v)
			case time.Time:
				newCell.Value = v.Format(time.RFC3339)
			case nil:
				newCell.Value = "NULL"
			case string:
				if v == "" {
					newCell.Value = "EMPTY"
				}
				newCell.Value = v
			default:
				newCell.Value = fmt.Sprintf("%v", v)
			}
			cells = append(cells, newCell)
		}
		rows = append(rows, cells)
	}

	if err := resultRows.Err(); err != nil {
		return nil, err
	}

	structure.Rows = rows

	// Get table indexes
	query = `
		SELECT
			i.relname                                  AS index_name,
			am.amname                                  AS index_algorithm,      -- btree/gin/gist/brin/hash
			idx.indisunique                            AS is_unique,
			idx.indisprimary                           AS is_primary,
			-- key columns (handles expressions too) – first (...) in the index def
			substring(pg_get_indexdef(idx.indexrelid) from '\(([^)]+)\)') AS columns,
			-- partial index predicate (NULL if not partial)
			pg_get_expr(idx.indpred, idx.indrelid)     AS condition,
			-- included (non-key) columns, if present (v11+); otherwise NULL
			substring(pg_get_indexdef(idx.indexrelid) from 'INCLUDE \(([^)]*)\)') AS include,
			obj_description(i.oid, 'pg_class')         AS comment
		FROM pg_class t
		JOIN pg_index idx ON t.oid = idx.indrelid
		JOIN pg_class i   ON i.oid = idx.indexrelid
		JOIN pg_am am     ON i.relam = am.oid
		WHERE t.relname = $1
		AND t.relnamespace = 'public'::regnamespace  -- adjust schema if needed
		ORDER BY i.relname DESC;
	`
	resultRows, err = pool.Query(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer resultRows.Close()

	var indexes model.Indexes

	fieldDescriptions = resultRows.FieldDescriptions()
	columns = make([]string, len(fieldDescriptions))
	for i, column := range fieldDescriptions {
		columns[i] = string(column.Name)
	}
	indexes.Columns = columns

	var indexRows [][]model.Cell

	for resultRows.Next() {
		row, err := resultRows.Values()
		if err != nil {
			return nil, err
		}

		cells := []model.Cell{}
		for i, cell := range row {
			newCell := model.Cell{
				Column: columns[i],
			}
			switch v := cell.(type) {
			case []byte:
				newCell.Value = string(v)
			case time.Time:
				newCell.Value = v.Format(time.RFC3339)
			case nil:
				newCell.Value = "NULL"
			case string:
				if v == "" {
					newCell.Value = "EMPTY"
				}
				newCell.Value = v
			default:
				newCell.Value = fmt.Sprintf("%v", v)
			}
			cells = append(cells, newCell)
		}
		indexRows = append(indexRows, cells)
	}

	if err := resultRows.Err(); err != nil {
		return nil, err
	}

	indexes.Rows = indexRows

	// Get table rules
	query = `
		SELECT
			con.conname        AS constraint_name,
			CASE con.contype
				WHEN 'p' THEN 'PRIMARY KEY'
				WHEN 'u' THEN 'UNIQUE'
				WHEN 'f' THEN 'FOREIGN KEY'
				WHEN 'c' THEN 'CHECK'
				WHEN 'x' THEN 'EXCLUSION'
				ELSE con.contype::text
			END                AS constraint_type,
			pg_get_constraintdef(con.oid) AS definition,
			con.convalidated   AS is_validated,
			obj_description(con.oid, 'pg_constraint') AS comment
		FROM pg_constraint con
		JOIN pg_class rel   ON rel.oid = con.conrelid
		JOIN pg_namespace n ON n.oid = rel.relnamespace
		WHERE rel.relname = $1
		AND n.nspname = 'public'   -- adjust schema if needed
		ORDER BY con.contype ASC;
	`
	resultRows, err = pool.Query(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer resultRows.Close()

	var rules model.Rules

	fieldDescriptions = resultRows.FieldDescriptions()
	columns = make([]string, len(fieldDescriptions))
	for i, column := range fieldDescriptions {
		columns[i] = string(column.Name)
	}
	rules.Columns = columns

	var ruleRows [][]model.Cell

	for resultRows.Next() {
		row, err := resultRows.Values()
		if err != nil {
			return nil, err
		}

		cells := []model.Cell{}
		for i, cell := range row {
			newCell := model.Cell{
				Column: columns[i],
			}
			switch v := cell.(type) {
			case []byte:
				newCell.Value = string(v)
			case time.Time:
				newCell.Value = v.Format(time.RFC3339)
			case nil:
				newCell.Value = "NULL"
			case string:
				if v == "" {
					newCell.Value = "EMPTY"
				}
				newCell.Value = v
			default:
				newCell.Value = fmt.Sprintf("%v", v)
			}
			cells = append(cells, newCell)
		}
		ruleRows = append(ruleRows, cells)
	}

	if err := resultRows.Err(); err != nil {
		return nil, err
	}

	rules.Rows = ruleRows

	return &model.TableInfo{Structure: structure, Indexes: indexes, Rules: rules}, nil
}
