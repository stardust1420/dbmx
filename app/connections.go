package app

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"dbmx/model"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

type Connections struct {
	DB *sql.DB
	PM *PoolManager

	// Track active queries per tab
	mu            sync.Mutex
	activeQueries map[queryKey]context.CancelCauseFunc
	querySeq      uint64

	// Map to save connection level table oid and names
	// Table Oid uniquely identifies a teble within a database. But it can repeat for a different database.
	// Hence use a nested map with connection uuid and table oid as key to store table name
	tableMu         sync.RWMutex
	tableOidNameMap map[uuid.UUID]map[uint32]string

	// Map to save connection level column type oid and their display info.
	// Type oids are per-database in the same way table oids are, so this is
	// keyed by pool id too. Entries are only ever added, never invalidated: a
	// type's name and category cannot change under a live oid.
	typeMu         sync.RWMutex
	typeOidInfoMap map[uuid.UUID]map[uint32]pgTypeInfo
}

// pgTypeInfo is what a result-set column's type oid resolves to: the name
// postgres itself prints for the type, the short name it holds the type under,
// and its catalog category.
type pgTypeInfo struct {
	Name        string
	DisplayName string
	Category    string
}

func NewConnections(db *sql.DB, pm *PoolManager) *Connections {
	return &Connections{
		DB:              db,
		PM:              pm,
		activeQueries:   make(map[queryKey]context.CancelCauseFunc),
		tableOidNameMap: make(map[uuid.UUID]map[uint32]string),
		typeOidInfoMap:  make(map[uuid.UUID]map[uint32]pgTypeInfo),
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

func (m *Connections) GetConnection(id int64) (model.Connection, error) {
	var connection model.Connection

	err := m.DB.QueryRow("SELECT * FROM connections WHERE id = ?", id).Scan(
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
		return connection, err
	}

	return connection, nil
}

func (m *Connections) UpdateConnection(c model.Connection) (bool, error) {
	// Check if the connection exists
	var exists bool
	err := m.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM connections WHERE id = ?)", c.ID).Scan(&exists)
	if err != nil {
		return false, err
	}

	if !exists {
		return false, errors.New("connection does not exist")
	}

	// Check if the connection is active
	if numPools, exists := m.PM.ActiveConns[c.ID]; exists {
		return false, errors.New(fmt.Sprintf("This connection has %d active instance(s). Please close all instances before updating", numPools))
	}

	// Update the connection
	_, err = m.DB.Exec("UPDATE connections SET engine = ?, host = ?, port = ?, username = ?, password = ?, database = ?, name = ?, env = ?, color = ?, is_advanced = ?, ssl_mode = ?, client_key = ?, client_cert = ?, root_ca_cert = ?, over_ssh = ?, ssh_host = ?, ssh_port = ?, ssh_username = ?, ssh_password = ?, use_ssh_key = ?, ssh_key = ? WHERE id = ?", c.Engine, c.Host, c.Port, c.Username, c.Password, c.Database, c.Name, c.Env, c.Color, c.IsAdvanced, c.SSLMode, c.ClientKey, c.ClientCert, c.RootCACert, c.OverSSH, c.SSHHost, c.SSHPort, c.SSHUsername, c.SSHPassword, c.UseSSHKey, c.SSHKey, c.ID)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (m *Connections) DeleteConnection(id int64) (bool, error) {
	// Check if the connection exists
	var exists bool
	err := m.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM connections WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		return false, err
	}

	if !exists {
		return false, errors.New("connection does not exist")
	}

	// Check if the connection is active
	if numPools, exists := m.PM.ActiveConns[id]; exists {
		return false, errors.New(fmt.Sprintf("This connection has %d active instance(s). Please close all instances before deleting", numPools))
	}

	// Delete the connection
	_, err = m.DB.Exec("DELETE FROM connections WHERE id = ?", id)
	if err != nil {
		return false, err
	}

	return true, nil
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

	if c.IsAdvanced && (len(c.RootCACert) > 0 || len(c.ClientKey) > 0 || len(c.ClientCert) > 0) {
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

	conn.Database = dbName

	cfg, err := BuildPostgresConnConfig(conn)
	if err != nil {
		return nil, err
	}

	activePoolID := uuid.New()

	// Establish connection and add pool to active pool manager
	_, err = c.PM.AddPool(activePoolID, cfg, id)
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

	// In case of table rows, find all the rows with type table where
	// active_db_id is null and postgres_connection_id and database matches
	// set the active pool id and active db properties in such tabs
	_, err = c.DB.Exec("UPDATE tabs SET active_db_id = ?, active_db = ?, active_db_color = ? WHERE connection_id = ? AND db_name = ?", activePoolID.String(), activeDB, conn.Color, id, dbName)
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
	_, err = c.PM.AddPool(activePoolID, cfg, id)
	if err != nil {
		return nil, err
	}

	activeDB := conn.Name + " - " + conn.Database

	// active_db_id is null and postgres_connection_id and database matches
	// set the active pool id and active db properties in such tabs
	_, err = c.DB.Exec("UPDATE tabs SET active_db_id = ?, active_db = ?, active_db_color = ? WHERE connection_id = ? AND db_name = ?", activePoolID.String(), activeDB, conn.Color, id, conn.Database)
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
	query := `
		SELECT 
			c.oid AS table_oid, 
			c.relname AS tablename
		FROM 
			pg_class c
		JOIN 
			pg_namespace n ON n.oid = c.relnamespace
		WHERE 
			n.nspname = 'public' 
			AND c.relkind IN ('r', 'p');
	`
	rows, err := pool.Query(context.TODO(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Slice to hold results
	var tables []string

	// Iterate through the rows
	for rows.Next() {
		var tableOID uint32
		var table string
		err := rows.Scan(&tableOID, &table)
		if err != nil {
			return nil, err
		}

		tables = append(tables, table)

		c.setTableOidNameMap(activePoolID, tableOID, table)
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

func (c *Connections) TerminatePostgresDatabaseConnection(activePoolID string, id int64) (bool, error) {
	activePoolIDUUID, err := uuid.Parse(activePoolID)
	if err != nil {
		return false, err
	}
	// Remove the db pool from active pools
	err = c.PM.DeletePool(activePoolIDUUID, id)
	if err != nil {
		return false, err
	}

	c.deleteTableOidNameMap(activePoolIDUUID)
	c.deleteTypeOidInfoMap(activePoolIDUUID)

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

		c.deleteTableOidNameMap(id)
		c.deleteTypeOidInfoMap(id)
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

func (c *Connections) ExecuteQuery(tabID int64, query string, isExplain bool) model.QueryResult {
	// Fetch Active Pool ID from Tab
	var activePoolID *string
	err := c.DB.QueryRow("SELECT active_db_id FROM tabs WHERE id = ?", tabID).Scan(&activePoolID)
	if err != nil {
		return model.QueryResult{OK: false, Message: "Tab doesn't exist"}
	}
	if activePoolID == nil {
		return model.QueryResult{OK: false, Message: "pool doesn't exist in tab in db"}
	}
	activePoolIDUUID, err := uuid.Parse(*activePoolID)
	if err != nil {
		return model.QueryResult{OK: false, Message: "invalid active pool id in tab"}
	}

	if isExplain {
		query = "EXPLAIN " + query
	}

	pool, exists := c.PM.GetPool(activePoolIDUUID)
	if !exists {
		return model.QueryResult{OK: false, Message: "pool doesn't exist"}
	}

	// --- 1. CONCURRENCY CONTROL & TIMEOUT SETUP ---
	ctx, cancel, release, err := c.beginQuery(tabID, queryKindEditor)
	if err != nil {
		return model.QueryResult{OK: false, Message: err.Error()}
	}
	defer release()
	// ----------------------------------------------

	response := model.QueryResult{OK: true}
	normalizedQuery := strings.ToLower(strings.TrimSpace(query))
	isWrite := isWriteOperation(normalizedQuery)

	startTime := time.Now()

	if isWrite {
		tag, err := pool.Exec(ctx, query)
		if err != nil {
			return c.handleQueryError(ctx, err)
		}
		response.RowsAffected = tag.RowsAffected()
		response.Columns = []string{"Rows Affected"}
		response.Rows = [][]model.Cell{{model.Cell{Column: "Rows Affected", Value: fmt.Sprintf("%d", response.RowsAffected)}}}
	} else {
		resultRows, err := pool.Query(ctx, query)
		if err != nil {
			return c.handleQueryError(ctx, err)
		}
		defer resultRows.Close()

		// Table oid set
		tableOidSet := make(map[uint32]struct{})
		idExists := false

		columns := resultRows.FieldDescriptions()
		columnNames := make([]string, len(columns))
		for i, column := range columns {
			columnName := string(column.Name)
			columnNames[i] = columnName
			tableOidSet[column.TableOID] = struct{}{}
			if columnName == "id" {
				idExists = true
			}
		}
		response.Columns = columnNames
		// Resolved before the rows are read, so a result that is cut short by the
		// timeout or the size cap still returns with its header types intact.
		response.ColumnTypes = c.columnTypesFor(ctx, activePoolIDUUID, pool, columns)

		// Set response table name if query output contains only one table data and has an id column
		if len(tableOidSet) == 1 && idExists {
			for oid := range tableOidSet {
				response.TableName = c.getTableOidNameMap(activePoolIDUUID, oid)
			}
		}

		var rows [][]model.Cell

		// Define your memory limit (e.g., 5 Megabytes)
		const maxAllowedBytes = 5 * 1024 * 1024
		var estimatedBytes int64

		for resultRows.Next() {
			// 1. ABORT CHECK: Did the user cancel or did the timeout hit during iteration?
			select {
			case <-ctx.Done():
				// Return partial results collected so far
				cancel()
				response.Rows = rows
				response.Message = partialResultMessage(ctx)
				response.ExecutionTime = time.Since(startTime).Milliseconds()
				return response
			default:
				// Context is still alive, proceed.
			}

			row, err := resultRows.Values()
			if err != nil {
				return model.QueryResult{OK: false, Message: err.Error()}
			}

			cells := make([]model.Cell, 0, len(row))
			for i, cell := range row {
				newCell := model.Cell{Column: columnNames[i]}

				// ... (Keep your existing switch statement to format newCell.Value) ...
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
					} else {
						newCell.Value = v
					}
				default:
					newCell.Value = fmt.Sprintf("%v", v)
				}

				// 2. MEMORY CHECK: Estimate the size of the cell we just created
				// We count the string length + ~32 bytes for struct/pointer overhead in Go
				estimatedBytes += int64(len(newCell.Value)) + 32
				cells = append(cells, newCell)
			}

			// Add slice overhead for the row (~24 bytes)
			estimatedBytes += 24
			rows = append(rows, cells)

			// If we exceed the limit, stop processing and return partial results
			if estimatedBytes > maxAllowedBytes {
				// Call cancel() to tell PostgreSQL to stop sending data over the network
				cancel()
				response.Rows = rows
				response.Message = "Result set exceeded 5MB limit. Partial results returned. Please add a LIMIT clause to your query."
				response.ExecutionTime = time.Since(startTime).Milliseconds()
				return response
			}
		}

		if err := resultRows.Err(); err != nil {
			// If the error is a timeout or cancellation and we have partial rows,
			// return the partial results with a warning instead of failing
			if (errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)) && len(rows) > 0 {
				response.Rows = rows
				response.Message = partialResultMessage(ctx)
				response.ExecutionTime = time.Since(startTime).Milliseconds()
				return response
			}
			return c.handleQueryError(ctx, err)
		}

		response.Rows = rows
	}

	// Save successful query to history
	if _, err := c.DB.Exec(`INSERT INTO query_history (query) VALUES (?)`, query); err != nil {
		log.Printf("failed to save query to history: %v", err)
	}

	response.ExecutionTime = time.Since(startTime).Milliseconds()
	return response
}

// queryTimeout caps how long a query may run before the server is asked to stop
// it. Generous on purpose: a slow report deserves the chance to finish, and a
// user who does not want to wait can stop it themselves.
const queryTimeout = 5 * time.Minute

var (
	// errQueryCancelledByUser is recorded as the cancellation cause when the user
	// presses stop. The cause is what separates a deliberate cancel from a
	// timeout, because the error pgx returns cannot: a cancel that reaches the
	// server comes back as postgres error 57014, which says only that somebody
	// cancelled the statement.
	errQueryCancelledByUser = errors.New("query cancelled by user")

	// errQueryTimedOut is the cause recorded when queryTimeout expires.
	errQueryTimedOut = errors.New("query timed out")

	errQueryAlreadyRunning = errors.New("a query is already running on this tab")
)

// queryKind separates the two things one tab can be running. They are tracked
// apart so that a table-view refresh and an editor run on the same tab do not
// block each other, while a single stop still reaches both.
type queryKind string

const (
	queryKindEditor queryKind = "editor"
	queryKindTable  queryKind = "table"
)

type queryKey struct {
	tabID int64
	kind  queryKind

	// seq distinguishes concurrent table loads. Paging through a table fires
	// them back to back and they have never excluded each other, so each gets
	// its own key. Editor queries leave it zero: there is only ever one, and
	// that is what makes a second run refusable.
	seq uint64
}

// beginQuery registers a cancellable context for one tab's query. It returns the
// context, a stop function for abandoning the result stream locally, and a
// release function that must be deferred to unregister the query.
func (c *Connections) beginQuery(tabID int64, kind queryKind) (context.Context, context.CancelFunc, func(), error) {
	key := queryKey{tabID: tabID, kind: kind}

	c.mu.Lock()
	if c.activeQueries == nil {
		c.activeQueries = make(map[queryKey]context.CancelCauseFunc)
	}
	if kind == queryKindEditor {
		if _, running := c.activeQueries[key]; running {
			c.mu.Unlock()
			return nil, nil, nil, errQueryAlreadyRunning
		}
	} else {
		c.querySeq++
		key.seq = c.querySeq
	}

	// Two layers, so that a timeout and a user cancel leave distinguishable
	// causes behind: the deadline sets errQueryTimedOut, CancelQuery sets
	// errQueryCancelledByUser, and Cause reports whichever fired.
	parent, cancelCause := context.WithCancelCause(context.Background())
	ctx, stop := context.WithTimeoutCause(parent, queryTimeout, errQueryTimedOut)

	c.activeQueries[key] = cancelCause
	c.mu.Unlock()

	release := func() {
		stop()
		cancelCause(nil)
		c.mu.Lock()
		delete(c.activeQueries, key)
		c.mu.Unlock()
	}
	return ctx, stop, release, nil
}

// CancelQuery stops whatever the tab is running, an editor query or a table-view
// load or both, and reports whether there was anything to stop.
func (c *Connections) CancelQuery(tabID int64) bool {
	c.mu.Lock()
	cancels := make([]context.CancelCauseFunc, 0, 2)
	for key, cancel := range c.activeQueries {
		if key.tabID == tabID {
			cancels = append(cancels, cancel)
		}
	}
	c.mu.Unlock()

	for _, cancel := range cancels {
		cancel(errQueryCancelledByUser)
	}
	return len(cancels) > 0
}

// partialResultMessage explains why a result set was cut short.
func partialResultMessage(ctx context.Context) string {
	if errors.Is(context.Cause(ctx), errQueryCancelledByUser) {
		return "Query cancelled. Partial results returned."
	}
	return fmt.Sprintf("Query timed out after %s. Partial results returned.", queryTimeout)
}

// Helper to handle standard vs timeout errors consistently
func (c *Connections) handleQueryError(ctx context.Context, err error) model.QueryResult {
	// Consult the cancellation cause before the error itself. A cancel that
	// actually reached the server arrives as postgres error 57014 rather than
	// context.Canceled, so the error on its own cannot say what stopped the query.
	switch cause := context.Cause(ctx); {
	case errors.Is(cause, errQueryCancelledByUser):
		return model.QueryResult{OK: false, Message: "Query cancelled"}
	case errors.Is(cause, errQueryTimedOut):
		return model.QueryResult{OK: false, Message: fmt.Sprintf("Query timed out after %s", queryTimeout)}
	}

	// Check if the error was caused by our context timing out or being canceled
	if errors.Is(err, context.DeadlineExceeded) {
		return model.QueryResult{OK: false, Message: "Query timed out after exceeding the maximum allowed time"}
	}
	if errors.Is(err, context.Canceled) {
		return model.QueryResult{OK: false, Message: "Query was manually aborted"}
	}

	// Your original fallback for syntax/SQL errors
	return model.QueryResult{
		OK:           true,
		Message:      err.Error(),
		RowsAffected: 0,
		Columns:      []string{"Error"},
		Rows:         [][]model.Cell{{model.Cell{Column: "Error", Value: err.Error()}}},
	}
}

// buildTableSelect renders the SELECT the table view runs, up to but not
// including its LIMIT/OFFSET. The export path shares it so that "export all
// rows" cannot drift from the filters the grid is actually paging through.
func buildTableSelect(tableName, selectQuery, where, orderBy, groupBy string) string {
	if strings.TrimSpace(selectQuery) == "" {
		selectQuery = "*"
	}

	query := fmt.Sprintf("SELECT %s FROM \"%s\"", selectQuery, tableName)
	if strings.TrimSpace(where) != "" {
		query += fmt.Sprintf(" WHERE %s", strings.TrimSpace(where))
	}
	if strings.TrimSpace(groupBy) != "" {
		query += fmt.Sprintf(" GROUP BY %s", strings.TrimSpace(groupBy))
	}
	if strings.TrimSpace(orderBy) != "" {
		query += fmt.Sprintf(" ORDER BY %s", strings.TrimSpace(orderBy))
	} else {
		query += " ORDER BY 1"
	}

	return query
}

func (c *Connections) GetTableData(tabID int64, tableName, selectQuery, limit, offset, where, orderBy, groupBy string, isPageData bool) model.QueryResult {
	// Fetch Active Pool ID from Tab
	var activePoolID *string
	err := c.DB.QueryRow("SELECT active_db_id FROM tabs WHERE id = ?", tabID).Scan(&activePoolID)
	if err != nil {
		return model.QueryResult{OK: false, Message: "Tab doesn't exist"}
	}
	if activePoolID == nil {
		return model.QueryResult{OK: false, Message: "pool doesn't exist in tab in db"}
	}
	activePoolIDUUID, err := uuid.Parse(*activePoolID)
	if err != nil {
		return model.QueryResult{OK: false, Message: "invalid active pool id in tab"}
	}

	pool, exists := c.PM.GetPool(activePoolIDUUID)
	if !exists {
		return model.QueryResult{OK: false, Message: "pool doesn't exist"}
	}

	ctx, _, release, err := c.beginQuery(tabID, queryKindTable)
	if err != nil {
		return model.QueryResult{OK: false, Message: err.Error()}
	}
	defer release()

	response := model.QueryResult{OK: true}

	setLimit := strconv.Itoa(20)
	if strings.TrimSpace(limit) != "" {
		limitInt, err := strconv.Atoi(strings.TrimSpace(limit))
		if err != nil {
			return model.QueryResult{OK: false, Message: "limit is not a number"}
		}
		if limitInt > 100 {
			return model.QueryResult{OK: false, Message: "limit cannot be greater than 100"}
		}
		setLimit = strings.TrimSpace(limit)
	}

	query := buildTableSelect(tableName, selectQuery, where, orderBy, groupBy)

	if !isPageData {
		// Get total rows count
		totalRowsQuery := fmt.Sprintf("SELECT COUNT(*) FROM \"%s\"", tableName)
		if strings.TrimSpace(where) != "" {
			totalRowsQuery += fmt.Sprintf(" WHERE %s", strings.TrimSpace(where))
		}
		if strings.TrimSpace(groupBy) != "" {
			totalRowsQuery += fmt.Sprintf(" GROUP BY %s", strings.TrimSpace(groupBy))
		}

		// Get total rows count
		var totalRows int64
		err := pool.QueryRow(ctx, totalRowsQuery).Scan(&totalRows)
		if err != nil {
			return c.handleQueryError(ctx, err)
		}
		response.TotalRows = totalRows
	}

	// Set limit
	query += fmt.Sprintf(" LIMIT %s", setLimit)

	if strings.TrimSpace(offset) != "" {
		query += fmt.Sprintf(" OFFSET %s", strings.TrimSpace(offset))
	}

	start := time.Now()

	// Use Query for read operations
	resultRows, err := pool.Query(ctx, query)
	if err != nil {
		return c.handleQueryError(ctx, err)
	}
	defer resultRows.Close()

	columns := resultRows.FieldDescriptions()
	columnNames := make([]string, len(columns))
	for i, column := range columns {
		columnNames[i] = string(column.Name)
	}
	response.Columns = columnNames
	response.ColumnTypes = c.columnTypesFor(ctx, activePoolIDUUID, pool, columns)

	var rows [][]model.Cell

	for resultRows.Next() {
		row, err := resultRows.Values()
		if err != nil {
			return model.QueryResult{OK: false, Message: err.Error()}
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
		return c.handleQueryError(ctx, err)
	}

	response.Rows = rows
	response.ExecutionTime = time.Since(start).Milliseconds()

	return response
}

func (c *Connections) GetTableInfo(tabID int64, tableName string) (*model.TableInfo, error) {
	// Fetch Active Pool ID from Tab
	var activePoolID *string
	err := c.DB.QueryRow("SELECT active_db_id FROM \"tabs\" WHERE id = ?", tabID).Scan(&activePoolID)
	if err != nil {
		return nil, errors.Wrap(err, "Tab doesn't exist")
	}
	if activePoolID == nil {
		return nil, errors.New("Active pool doesn't exist in tab")
	}
	activePoolIDUUID, err := uuid.Parse(*activePoolID)
	if err != nil {
		return nil, errors.Wrap(err, "Invalid active pool in tab")
	}

	pool, exists := c.PM.GetPool(activePoolIDUUID)
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
			obj_description(i.oid, 'pg_class')         AS comment,
			-- an index postgres created for a constraint is owned by that constraint
			EXISTS (SELECT 1 FROM pg_constraint c WHERE c.conindid = idx.indexrelid) AS is_constraint
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
				WHEN 'x' THEN 'EXCLUDE'
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

func (c *Connections) UpdateCells(tabID int64, updateCells []model.UpdateCell) (bool, error) {
	// Fetch Active Pool ID from Tab
	var activePoolID *string
	err := c.DB.QueryRow("SELECT active_db_id FROM tabs WHERE id = ?", tabID).Scan(&activePoolID)
	if err != nil {
		return false, errors.Wrap(err, "Tab doesn't exist")
	}
	if activePoolID == nil {
		return false, errors.New("Active pool doesn't exist in tab")
	}
	activePoolIDUUID, err := uuid.Parse(*activePoolID)
	if err != nil {
		return false, errors.Wrap(err, "Invalid active pool in tab")
	}

	pool, exists := c.PM.GetPool(activePoolIDUUID)
	if !exists {
		return false, errors.New("pool doesn't exist")
	}

	// 1. Initialize a new batch
	batch := &pgx.Batch{}

	for _, u := range updateCells {
		// 2. SAFELY construct the query using pgx.Identifier for table/column names.
		// This translates "my_table" to `"my_table"` and prevents SQL injection.
		safeTable := pgx.Identifier{u.TableName}.Sanitize()
		safeColumn := pgx.Identifier{u.ColumnName}.Sanitize()

		// Construct the final query string
		query := fmt.Sprintf("UPDATE %s SET %s = $1 WHERE id = $2", safeTable, safeColumn)

		// 3. Queue the query with the actual parameterized values
		batch.Queue(query, u.Value, u.RowID)
	}

	// 4. Send the batch to the database
	br := pool.SendBatch(context.Background(), batch)

	// 5. CRITICAL: You must ensure the batch results are closed to release the connection
	defer br.Close()

	// 6. Verify the results. You must call Exec() (or QueryRow) for EVERY item you queued.
	for i := 0; i < len(updateCells); i++ {
		_, err := br.Exec()
		if err != nil {
			// If one fails, the whole batch transaction rolls back automatically
			return false, fmt.Errorf("failed to update cell %s (table: %s, row: %d): %w",
				updateCells[i].CellID, updateCells[i].TableName, updateCells[i].RowID, err)
		}
	}

	return true, nil
}

func (c *Connections) setTableOidNameMap(activePoolID uuid.UUID, tableOid uint32, tableName string) {
	c.tableMu.Lock()
	defer c.tableMu.Unlock()
	if c.tableOidNameMap[activePoolID] == nil {
		c.tableOidNameMap[activePoolID] = make(map[uint32]string)
	}
	c.tableOidNameMap[activePoolID][tableOid] = tableName
}

func (c *Connections) getTableOidNameMap(activePoolID uuid.UUID, tableOid uint32) string {
	c.tableMu.RLock()
	defer c.tableMu.RUnlock()
	if c.tableOidNameMap[activePoolID] == nil {
		return ""
	}
	return c.tableOidNameMap[activePoolID][tableOid]
}

func (c *Connections) deleteTableOidNameMap(activePoolID uuid.UUID) {
	c.tableMu.Lock()
	defer c.tableMu.Unlock()
	delete(c.tableOidNameMap, activePoolID)
}

func (c *Connections) deleteTypeOidInfoMap(activePoolID uuid.UUID) {
	c.typeMu.Lock()
	defer c.typeMu.Unlock()
	delete(c.typeOidInfoMap, activePoolID)
}

// format_type prints a type the way postgres itself writes it in DDL -- "integer"
// rather than "int4", "text[]" rather than "_text" -- and resolves a domain or an
// extension type by the same rule, so no list of known type names is needed here.
// The modifier is passed as NULL deliberately: a modifier belongs to the column,
// not the type, and leaving it off is what makes one row per oid cacheable.
// typcategory is postgres's own single-byte "char" type, which has no binary
// decoding into a Go string, so it is cast on the server side.
//
// typname is the short name -- "int4", "varchar", "timestamptz" -- except for an
// array, which postgres names by prefixing an underscore to its element type.
// "_text" is not a name anyone reads, so an array is printed instead.
const columnTypeOidQuery = `
	SELECT
		t.oid,
		format_type(t.oid, NULL),
		CASE WHEN t.typcategory = 'A' THEN format_type(t.oid, NULL) ELSE t.typname END,
		t.typcategory::text
	FROM pg_type t
	WHERE t.oid = ANY($1::oid[])`

// columnTypesFor resolves the type of every column of a result set, for the type
// icons the grid draws in its header.
//
// Resolution is best-effort. The rows of the query the user actually ran matter
// more than the icons above them, so a failure here is logged and leaves the
// types empty; the grid falls back to a plain header.
func (c *Connections) columnTypesFor(ctx context.Context, activePoolID uuid.UUID, pool *pgxpool.Pool, fields []pgconn.FieldDescription) []model.ColumnType {
	columnTypes := make([]model.ColumnType, len(fields))

	// Collect the oids this pool has not resolved before. Every query after the
	// first over the same types answers from the cache alone.
	unresolved := []uint32{}
	c.typeMu.RLock()
	cached := c.typeOidInfoMap[activePoolID]
	for i, field := range fields {
		columnTypes[i] = model.ColumnType{Name: field.Name}
		if info, ok := cached[field.DataTypeOID]; ok {
			columnTypes[i].DataType = info.Name
			columnTypes[i].DisplayType = info.DisplayName
			columnTypes[i].Category = info.Category
			continue
		}
		unresolved = append(unresolved, field.DataTypeOID)
	}
	c.typeMu.RUnlock()

	if len(unresolved) == 0 {
		c.applyColumnAttributes(ctx, pool, fields, columnTypes)
		return columnTypes
	}

	rows, err := pool.Query(ctx, columnTypeOidQuery, unresolved)
	if err != nil {
		log.Printf("failed to resolve column type oids: %v", err)
		c.applyColumnAttributes(ctx, pool, fields, columnTypes)
		return columnTypes
	}
	defer rows.Close()

	resolved := make(map[uint32]pgTypeInfo, len(unresolved))
	for rows.Next() {
		var oid uint32
		var info pgTypeInfo
		if err := rows.Scan(&oid, &info.Name, &info.DisplayName, &info.Category); err != nil {
			log.Printf("failed to scan column type oid: %v", err)
			c.applyColumnAttributes(ctx, pool, fields, columnTypes)
			return columnTypes
		}
		resolved[oid] = info
	}
	if err := rows.Err(); err != nil {
		log.Printf("failed to read column type oids: %v", err)
		c.applyColumnAttributes(ctx, pool, fields, columnTypes)
		return columnTypes
	}

	c.typeMu.Lock()
	if c.typeOidInfoMap[activePoolID] == nil {
		c.typeOidInfoMap[activePoolID] = make(map[uint32]pgTypeInfo, len(resolved))
	}
	for oid, info := range resolved {
		c.typeOidInfoMap[activePoolID][oid] = info
	}
	c.typeMu.Unlock()

	for i, field := range fields {
		if info, ok := resolved[field.DataTypeOID]; ok {
			columnTypes[i].DataType = info.Name
			columnTypes[i].DisplayType = info.DisplayName
			columnTypes[i].Category = info.Category
		}
	}

	c.applyColumnAttributes(ctx, pool, fields, columnTypes)

	return columnTypes
}

// A result column that is a plain reference to a table column carries the oid of
// that table and the column's position in it, which is enough to read the column's
// own rules out of the catalog. An expression, an aggregate or a literal carries
// no table oid at all and is skipped.
//
// Unlike a type name, none of this is cached: the app's own schema editor can drop
// a constraint or a NOT NULL out from under a grid that is still open, and a stale
// "PK" in a header is worse than the round trip it saves.
const columnAttributeQuery = `
	SELECT
		a.attrelid,
		a.attnum,
		NOT a.attnotnull,
		pk.conkey IS NOT NULL,
		COALESCE(cardinality(pk.conkey), 0) > 1,
		fk.confrelid::regclass::text
	FROM unnest($1::oid[], $2::smallint[]) AS t(attrelid, attnum)
	JOIN pg_attribute a ON a.attrelid = t.attrelid AND a.attnum = t.attnum
	LEFT JOIN LATERAL (
		SELECT c.conkey
		FROM pg_constraint c
		WHERE c.conrelid = a.attrelid AND c.contype = 'p' AND a.attnum = ANY(c.conkey)
		LIMIT 1
	) pk ON true
	LEFT JOIN LATERAL (
		SELECT c.confrelid
		FROM pg_constraint c
		WHERE c.conrelid = a.attrelid AND c.contype = 'f' AND a.attnum = ANY(c.conkey)
		ORDER BY c.conname
		LIMIT 1
	) fk ON true`

// columnAttribute keys one row of the result above back to the field that asked
// for it. A column can be selected twice in one query, so the pair is a lookup
// key rather than a position.
type columnAttributeKey struct {
	TableOID uint32
	AttNum   int16
}

// applyColumnAttributes fills in the key and nullability flags of every result
// column that has a table column behind it. Like type resolution it is
// best-effort: a failure leaves the flags unset and the header simply shows the
// column's type without its badges.
func (c *Connections) applyColumnAttributes(ctx context.Context, pool *pgxpool.Pool, fields []pgconn.FieldDescription, columnTypes []model.ColumnType) {
	tableOIDs := []uint32{}
	attNums := []int16{}
	for _, field := range fields {
		if field.TableOID == 0 {
			continue
		}
		tableOIDs = append(tableOIDs, field.TableOID)
		attNums = append(attNums, int16(field.TableAttributeNumber))
	}
	if len(tableOIDs) == 0 {
		return
	}

	rows, err := pool.Query(ctx, columnAttributeQuery, tableOIDs, attNums)
	if err != nil {
		log.Printf("failed to resolve column attributes: %v", err)
		return
	}
	defer rows.Close()

	attributes := make(map[columnAttributeKey]model.ColumnType, len(tableOIDs))
	for rows.Next() {
		var key columnAttributeKey
		var attribute model.ColumnType
		var foreignKeyTable *string
		if err := rows.Scan(
			&key.TableOID,
			&key.AttNum,
			&attribute.IsNullable,
			&attribute.IsPrimaryKey,
			&attribute.IsCompositeKey,
			&foreignKeyTable,
		); err != nil {
			log.Printf("failed to scan column attribute: %v", err)
			return
		}
		if foreignKeyTable != nil {
			attribute.IsForeignKey = true
			attribute.ForeignKeyTable = *foreignKeyTable
		}
		attributes[key] = attribute
	}
	if err := rows.Err(); err != nil {
		log.Printf("failed to read column attributes: %v", err)
		return
	}

	for i, field := range fields {
		attribute, ok := attributes[columnAttributeKey{
			TableOID: field.TableOID,
			AttNum:   int16(field.TableAttributeNumber),
		}]
		if !ok {
			continue
		}
		columnTypes[i].IsNullable = attribute.IsNullable
		columnTypes[i].IsPrimaryKey = attribute.IsPrimaryKey
		columnTypes[i].IsCompositeKey = attribute.IsCompositeKey
		columnTypes[i].IsForeignKey = attribute.IsForeignKey
		columnTypes[i].ForeignKeyTable = attribute.ForeignKeyTable
	}
}

// poolForTab resolves the live pgx pool a tab is currently pointed at.
func (c *Connections) poolForTab(tabID int64) (*pgxpool.Pool, error) {
	var activePoolID *string
	if err := c.DB.QueryRow("SELECT active_db_id FROM tabs WHERE id = ?", tabID).Scan(&activePoolID); err != nil {
		return nil, errors.Wrap(err, "Tab doesn't exist")
	}
	if activePoolID == nil {
		return nil, errors.New("Active pool doesn't exist in tab")
	}
	activePoolIDUUID, err := uuid.Parse(*activePoolID)
	if err != nil {
		return nil, errors.Wrap(err, "Invalid active pool in tab")
	}
	pool, exists := c.PM.GetPool(activePoolIDUUID)
	if !exists {
		return nil, errors.New("pool doesn't exist")
	}
	return pool, nil
}

const tableColumnsMetaQuery = `
	SELECT
		a.attname AS name,
		format_type(a.atttypid, a.atttypmod) AS data_type,
		format_type(a.atttypid, NULL) AS cast_type,
		t.typcategory::text AS category,
		NOT a.attnotnull AS is_nullable,
		pg_get_expr(d.adbin, d.adrelid) AS default_value,
		(a.attidentity = 'a' OR a.attgenerated <> '') AS is_read_only,
		COALESCE(i.indisprimary, false) AS is_primary_key,
		CASE WHEN t.typtype = 'e' THEN ARRAY(
			SELECT e.enumlabel::text FROM pg_enum e
			WHERE e.enumtypid = t.oid ORDER BY e.enumsortorder
		) END AS enum_values,
		col_description(a.attrelid, a.attnum) AS comment
	FROM pg_attribute a
	JOIN pg_class c ON c.oid = a.attrelid
	JOIN pg_namespace n ON n.oid = c.relnamespace
	JOIN pg_type t ON t.oid = a.atttypid
	LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
	LEFT JOIN pg_index i ON i.indrelid = c.oid AND i.indisprimary AND a.attnum = ANY(i.indkey)
	WHERE n.nspname = 'public'
		AND c.relname = $1
		AND a.attnum > 0
		AND NOT a.attisdropped
	ORDER BY a.attnum
`

// GetTableColumnsMeta returns the column definitions of a table in the tab's active
// database, ordered as declared. It drives the "add new row" form.
func (c *Connections) GetTableColumnsMeta(tabID int64, tableName string) ([]model.ColumnMeta, error) {
	pool, err := c.poolForTab(tabID)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(context.Background(), tableColumnsMetaQuery, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := []model.ColumnMeta{}
	for rows.Next() {
		var (
			col          model.ColumnMeta
			defaultValue *string
			comment      *string
			enumValues   []string
		)
		if err := rows.Scan(
			&col.Name,
			&col.DataType,
			&col.CastType,
			&col.Category,
			&col.IsNullable,
			&defaultValue,
			&col.IsReadOnly,
			&col.IsPrimaryKey,
			&enumValues,
			&comment,
		); err != nil {
			return nil, err
		}
		if defaultValue != nil {
			col.HasDefault = true
			col.DefaultValue = *defaultValue
		}
		if comment != nil {
			col.Comment = *comment
		}
		col.EnumValues = enumValues
		columns = append(columns, col)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(columns) == 0 {
		return nil, errors.Errorf("table %q not found", tableName)
	}

	return columns, nil
}

// InsertRow inserts a single row into tableName. Every value is bound as a plain
// parameter carrying its text form, so postgres infers the parameter type from the
// target column and parses the text with that column's own input function. That is
// the same path a hand-written INSERT takes, which means every type works without
// special casing -- json, arrays, enums, uuid, inet, bytea, ranges, user-defined --
// and length/precision are enforced by the column rather than by a cast. An explicit
// cast must not be used here: $n::text::bit resolves to bit(1) and $n::text::character
// to character(1), which silently mangle values destined for bit(4)/character(3), and
// casting to the modified type instead silently truncates varchar and zero-pads bit.
//
// Columns absent from values keep their database default, and a nil Value inserts SQL NULL.
func (c *Connections) InsertRow(tabID int64, tableName string, values []model.InsertValue) (bool, error) {
	pool, err := c.poolForTab(tabID)
	if err != nil {
		return false, err
	}

	columnsMeta, err := c.GetTableColumnsMeta(tabID, tableName)
	if err != nil {
		return false, err
	}
	metaByName := make(map[string]model.ColumnMeta, len(columnsMeta))
	for _, col := range columnsMeta {
		metaByName[col.Name] = col
	}

	// Validate every supplied column against the catalog. Only sanitized identifiers
	// reach the query text; the values themselves are always bound as parameters.
	var (
		columnList   []string
		placeholders []string
		args         []any
		seen         = make(map[string]bool, len(values))
	)
	for _, v := range values {
		col, ok := metaByName[v.ColumnName]
		if !ok {
			return false, errors.Errorf("column %q doesn't exist on table %q", v.ColumnName, tableName)
		}
		if col.IsReadOnly {
			return false, errors.Errorf("column %q is generated and cannot be set", v.ColumnName)
		}
		if seen[v.ColumnName] {
			return false, errors.Errorf("column %q supplied more than once", v.ColumnName)
		}
		seen[v.ColumnName] = true

		columnList = append(columnList, pgx.Identifier{col.Name}.Sanitize())
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)+1))
		args = append(args, v.Value)
	}

	safeTable := pgx.Identifier{tableName}.Sanitize()
	query := fmt.Sprintf("INSERT INTO %s DEFAULT VALUES", safeTable)
	if len(columnList) > 0 {
		query = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s)",
			safeTable,
			strings.Join(columnList, ", "),
			strings.Join(placeholders, ", "),
		)
	}

	tag, err := pool.Exec(context.Background(), query, args...)
	if err != nil {
		return false, err
	}

	return tag.RowsAffected() > 0, nil
}

// DeleteRows deletes the rows of tableName whose primary key "id" is in rowIDs, and
// returns how many were actually removed. Like the inline cell editor and UpdateCells,
// it identifies a row by a column literally named "id" -- that is the grid's contract,
// and the catalog lookup below turns a table without one into a clear error rather than
// a raw postgres message.
//
// The ids arrive as text and are bound as plain parameters so postgres infers each
// parameter's type from "id" itself, the same reasoning as InsertRow: an integer,
// uuid, or text primary key all work without special casing, and the comparison stays
// indexable (unlike casting the column with id::text).
func (c *Connections) DeleteRows(tabID int64, tableName string, rowIDs []string) (int64, error) {
	if len(rowIDs) == 0 {
		return 0, errors.New("no rows selected")
	}

	pool, err := c.poolForTab(tabID)
	if err != nil {
		return 0, err
	}

	// Validates that the table exists and resolves its columns.
	columnsMeta, err := c.GetTableColumnsMeta(tabID, tableName)
	if err != nil {
		return 0, err
	}
	hasID := false
	for _, col := range columnsMeta {
		if col.Name == "id" {
			hasID = true
			break
		}
	}
	if !hasID {
		return 0, errors.Errorf("table %q has no \"id\" column, so rows cannot be deleted from the grid", tableName)
	}

	placeholders := make([]string, 0, len(rowIDs))
	args := make([]any, 0, len(rowIDs))
	for _, id := range rowIDs {
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)+1))
		args = append(args, id)
	}

	query := fmt.Sprintf(
		"DELETE FROM %s WHERE id IN (%s)",
		pgx.Identifier{tableName}.Sanitize(),
		strings.Join(placeholders, ", "),
	)

	tag, err := pool.Exec(context.Background(), query, args...)
	if err != nil {
		return 0, err
	}

	return tag.RowsAffected(), nil
}

// The DDL queries all identify the table by its regclass, so the table name is quoted
// once into a schema-qualified literal and bound as a parameter from there on.
const tableDDLColumnsQuery = `
	SELECT
		a.attname,
		format_type(a.atttypid, a.atttypmod),
		a.attnotnull,
		pg_get_expr(ad.adbin, ad.adrelid),
		a.attidentity::text,
		a.attgenerated::text,
		col_description(a.attrelid, a.attnum)
	FROM pg_attribute a
	LEFT JOIN pg_attrdef ad ON ad.adrelid = a.attrelid AND ad.adnum = a.attnum
	WHERE a.attrelid = $1::regclass
		AND a.attnum > 0
		AND NOT a.attisdropped
	ORDER BY a.attnum
`

// Ordered so the primary key leads, then uniques, checks and finally foreign keys.
const tableDDLConstraintsQuery = `
	SELECT conname, pg_get_constraintdef(oid)
	FROM pg_constraint
	WHERE conrelid = $1::regclass
	ORDER BY
		CASE contype WHEN 'p' THEN 0 WHEN 'u' THEN 1 WHEN 'c' THEN 2 WHEN 'f' THEN 3 ELSE 4 END,
		conname
`

// Indexes that back a constraint are excluded: they are already emitted as part of the
// CREATE TABLE body, and re-issuing them here would be invalid.
const tableDDLIndexesQuery = `
	SELECT pg_get_indexdef(i.indexrelid)
	FROM pg_index i
	JOIN pg_class ic ON ic.oid = i.indexrelid
	WHERE i.indrelid = $1::regclass
		AND NOT EXISTS (SELECT 1 FROM pg_constraint c WHERE c.conindid = i.indexrelid)
	ORDER BY ic.relname
`

// quoteSQLLiteral renders s as a single-quoted SQL string literal.
func quoteSQLLiteral(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// GetTableDDL reconstructs a CREATE TABLE statement for tableName in the tab's active
// database, followed by its non-constraint indexes and any table or column comments.
//
// Constraints and indexes are rendered by postgres itself through pg_get_constraintdef
// and pg_get_indexdef rather than being assembled from information_schema, so partial
// indexes, expression indexes, operator classes, ON DELETE actions and check expressions
// all come out exactly as the server holds them. Column types likewise come from
// format_type, which keeps the type modifier -- character varying(255), numeric(10,2).
//
// The result is for reading and copying, not for round-tripping a schema: it deliberately
// omits ownership, grants, triggers, storage parameters and tablespaces.
func (c *Connections) GetTableDDL(tabID int64, tableName string) (string, error) {
	pool, err := c.poolForTab(tabID)
	if err != nil {
		return "", err
	}

	safeTable := pgx.Identifier{"public", tableName}.Sanitize()
	ctx := context.Background()

	// Columns.
	rows, err := pool.Query(ctx, tableDDLColumnsQuery, safeTable)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var (
		columnLines  []string
		columnNotes  []string
		columnsFound bool
	)
	for rows.Next() {
		var (
			name        string
			dataType    string
			notNull     bool
			defaultExpr *string
			identity    string
			generated   string
			comment     *string
		)
		if err := rows.Scan(&name, &dataType, &notNull, &defaultExpr, &identity, &generated, &comment); err != nil {
			return "", err
		}
		columnsFound = true

		line := fmt.Sprintf("    %s %s", pgx.Identifier{name}.Sanitize(), dataType)
		switch {
		case generated == "s" && defaultExpr != nil:
			// A stored generated column keeps its expression in pg_attrdef, but it is a
			// GENERATED clause rather than a DEFAULT and must be written as one.
			line += fmt.Sprintf(" GENERATED ALWAYS AS (%s) STORED", *defaultExpr)
		case defaultExpr != nil:
			line += " DEFAULT " + *defaultExpr
		}
		if notNull {
			line += " NOT NULL"
		}
		switch identity {
		case "a":
			line += " GENERATED ALWAYS AS IDENTITY"
		case "d":
			line += " GENERATED BY DEFAULT AS IDENTITY"
		}
		columnLines = append(columnLines, line)

		if comment != nil && *comment != "" {
			columnNotes = append(columnNotes, fmt.Sprintf(
				"COMMENT ON COLUMN %s.%s IS %s;",
				safeTable, pgx.Identifier{name}.Sanitize(), quoteSQLLiteral(*comment),
			))
		}
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	if !columnsFound {
		return "", errors.Errorf("table %q not found", tableName)
	}

	// Constraints, appended to the CREATE TABLE body.
	constraintRows, err := pool.Query(ctx, tableDDLConstraintsQuery, safeTable)
	if err != nil {
		return "", err
	}
	defer constraintRows.Close()

	bodyLines := columnLines
	for constraintRows.Next() {
		var name, definition string
		if err := constraintRows.Scan(&name, &definition); err != nil {
			return "", err
		}
		bodyLines = append(bodyLines, fmt.Sprintf(
			"    CONSTRAINT %s %s", pgx.Identifier{name}.Sanitize(), definition,
		))
	}
	if err := constraintRows.Err(); err != nil {
		return "", err
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "CREATE TABLE %s (\n%s\n);\n", safeTable, strings.Join(bodyLines, ",\n"))

	// Standalone indexes.
	indexRows, err := pool.Query(ctx, tableDDLIndexesQuery, safeTable)
	if err != nil {
		return "", err
	}
	defer indexRows.Close()

	var indexDefs []string
	for indexRows.Next() {
		var definition string
		if err := indexRows.Scan(&definition); err != nil {
			return "", err
		}
		indexDefs = append(indexDefs, definition+";")
	}
	if err := indexRows.Err(); err != nil {
		return "", err
	}
	if len(indexDefs) > 0 {
		sb.WriteString("\n" + strings.Join(indexDefs, "\n") + "\n")
	}

	// Comments.
	var tableComment *string
	if err := pool.QueryRow(ctx, "SELECT obj_description($1::regclass, 'pg_class')", safeTable).Scan(&tableComment); err != nil {
		return "", err
	}
	var notes []string
	if tableComment != nil && *tableComment != "" {
		notes = append(notes, fmt.Sprintf("COMMENT ON TABLE %s IS %s;", safeTable, quoteSQLLiteral(*tableComment)))
	}
	notes = append(notes, columnNotes...)
	if len(notes) > 0 {
		sb.WriteString("\n" + strings.Join(notes, "\n") + "\n")
	}

	return sb.String(), nil
}
