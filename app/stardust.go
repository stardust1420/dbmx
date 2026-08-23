package app

import (
	"context"
	"database/sql"
	"dbmx/config/env"
	"dbmx/model"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/stardust1420/dbmx-go"
)

type Stardust struct {
	DB   *sql.DB
	Env  *env.Env
	Auth *Auth
	PM   *PoolManager
}

func NewStardust(db *sql.DB, env *env.Env, auth *Auth, pm *PoolManager) *Stardust {
	return &Stardust{
		DB:   db,
		Env:  env,
		Auth: auth,
		PM:   pm,
	}
}

func (s *Stardust) EnableStardustAI() (dbmx.Customer, error) {
	// Fetch the active session from db
	token, err := s.Auth.getToken()
	if err != nil {
		return dbmx.Customer{}, err
	}

	dbmxClient := dbmx.NewClient(dbmx.Credentials{
		BaseURL:     s.Env.DBMXConfig.BaseURL,
		AccessToken: token.AccessToken,
	})

	customer, err := dbmxClient.EnableStardustAI(context.TODO())
	if err != nil {
		return dbmx.Customer{}, err
	}

	return customer, nil
}

func (s *Stardust) DisableStardustAI() error {
	// Fetch the active session from db
	token, err := s.Auth.getToken()
	if err != nil {
		return err
	}

	dbmxClient := dbmx.NewClient(dbmx.Credentials{
		BaseURL:     s.Env.DBMXConfig.BaseURL,
		AccessToken: token.AccessToken,
	})

	success, err := dbmxClient.DisableStardustAI(context.TODO())
	if err != nil {
		return err
	}
	if !success {
		return errors.New("Failed to disable stardust AI")
	}

	return nil
}

func (s *Stardust) SwitchDefaultKey(switchValue bool) error {
	// Fetch the active session from db
	token, err := s.Auth.getToken()
	if err != nil {
		return err
	}

	dbmxClient := dbmx.NewClient(dbmx.Credentials{
		BaseURL:     s.Env.DBMXConfig.BaseURL,
		AccessToken: token.AccessToken,
	})

	success, err := dbmxClient.SwitchDefaultKey(context.TODO(), switchValue)
	if err != nil {
		return err
	}
	if !success {
		return errors.New("Failed to switch default key")
	}

	return nil
}

func (s *Stardust) ListProviders() ([]dbmx.UserProvider, error) {
	// Fetch the active session from db
	token, err := s.Auth.getToken()
	if err != nil {
		return nil, err
	}

	dbmxClient := dbmx.NewClient(dbmx.Credentials{
		BaseURL:     s.Env.DBMXConfig.BaseURL,
		AccessToken: token.AccessToken,
	})

	userProviders, err := dbmxClient.ListProviders(context.TODO())
	if err != nil {
		return nil, err
	}

	return userProviders, nil
}

func (s *Stardust) AddProviderAPIKey(provider, apiKey string) error {
	// Fetch the active session from db
	token, err := s.Auth.getToken()
	if err != nil {
		return err
	}

	dbmxClient := dbmx.NewClient(dbmx.Credentials{
		BaseURL:     s.Env.DBMXConfig.BaseURL,
		AccessToken: token.AccessToken,
	})

	_, err = dbmxClient.AddProviderAPIKey(context.TODO(), dbmx.AddProviderAPIKeyReq{
		Provider: provider,
		APIKey:   apiKey,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Stardust) UpdateProviderAPIKey(keyID, provider, apiKey string) error {
	// Fetch the active session from db
	token, err := s.Auth.getToken()
	if err != nil {
		return err
	}

	dbmxClient := dbmx.NewClient(dbmx.Credentials{
		BaseURL:     s.Env.DBMXConfig.BaseURL,
		AccessToken: token.AccessToken,
	})

	_, err = dbmxClient.UpdateProviderAPIKey(context.TODO(), dbmx.UpdateProviderAPIKeyReq{
		KeyID:    keyID,
		Provider: provider,
		APIKey:   apiKey,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Stardust) DeleteProviderAPIKey(keyID, provider string) error {
	// Fetch the active session from db
	token, err := s.Auth.getToken()
	if err != nil {
		return err
	}

	dbmxClient := dbmx.NewClient(dbmx.Credentials{
		BaseURL:     s.Env.DBMXConfig.BaseURL,
		AccessToken: token.AccessToken,
	})

	_, err = dbmxClient.DeleteProviderAPIKey(context.TODO(), dbmx.DeleteProviderAPIKeyReq{
		KeyID:    keyID,
		Provider: provider,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Stardust) ListAvailableModels() ([]dbmx.Model, error) {
	// Fetch the active session from db
	token, err := s.Auth.getToken()
	if err != nil {
		return nil, err
	}

	dbmxClient := dbmx.NewClient(dbmx.Credentials{
		BaseURL:     s.Env.DBMXConfig.BaseURL,
		AccessToken: token.AccessToken,
	})

	response := []dbmx.Model{}

	availableModels, err := dbmxClient.ListAvailableModels(context.TODO())
	for _, model := range availableModels {
		if strings.TrimSpace(model.NormalizedName) != "" {
			response = append(response, model)
		}
	}

	return response, nil
}

func (s Stardust) GetSchemaContext(pool *pgxpool.Pool) (string, error) {
	query := `
		WITH columns_agg AS (
			SELECT
				c.table_name,
				string_agg(
					'  ' || c.column_name || ' ' ||
					UPPER(c.data_type) ||
					CASE WHEN c.character_maximum_length IS NOT NULL
						THEN '(' || c.character_maximum_length || ')' ELSE '' END ||
					CASE WHEN c.is_nullable = 'NO' THEN ' NOT NULL' ELSE '' END ||
					CASE WHEN c.column_default IS NOT NULL
						THEN ' DEFAULT ' || c.column_default ELSE '' END,
					E',\n'
					ORDER BY c.ordinal_position
				) AS col_defs
			FROM information_schema.columns c
			WHERE c.table_schema = 'public'
			GROUP BY c.table_name
		),
		constraints_agg AS (
			SELECT
				tc.table_name,
				string_agg(
					'  ' ||
					CASE tc.constraint_type
						WHEN 'PRIMARY KEY' THEN
							'PRIMARY KEY (' || (
								SELECT string_agg(kcu.column_name, ', ' ORDER BY kcu.ordinal_position)
								FROM information_schema.key_column_usage kcu
								WHERE kcu.constraint_name = tc.constraint_name
								AND kcu.table_schema = 'public'
							) || ')'
						WHEN 'FOREIGN KEY' THEN
							'FOREIGN KEY (' || (
								SELECT string_agg(kcu.column_name, ', ' ORDER BY kcu.ordinal_position)
								FROM information_schema.key_column_usage kcu
								WHERE kcu.constraint_name = tc.constraint_name
								AND kcu.table_schema = 'public'
							) || ') REFERENCES ' || (
								SELECT ccu.table_name || '(' || string_agg(ccu.column_name, ', ') || ')'
								FROM information_schema.constraint_column_usage ccu
								WHERE ccu.constraint_name = tc.constraint_name
								AND ccu.table_schema = 'public'
								GROUP BY ccu.table_name
							)
						WHEN 'UNIQUE' THEN
							'UNIQUE (' || (
								SELECT string_agg(kcu.column_name, ', ' ORDER BY kcu.ordinal_position)
								FROM information_schema.key_column_usage kcu
								WHERE kcu.constraint_name = tc.constraint_name
								AND kcu.table_schema = 'public'
							) || ')'
					END,
					E',\n'
				) AS constraint_defs
			FROM information_schema.table_constraints tc
			WHERE tc.table_schema = 'public'
			AND tc.constraint_type IN ('PRIMARY KEY','FOREIGN KEY','UNIQUE')
			GROUP BY tc.table_name
		)
		SELECT
			'CREATE TABLE ' || ca.table_name || ' (\n' ||
			ca.col_defs ||
			COALESCE(E',\n' || co.constraint_defs, '') ||
			E'\n);' AS ddl
		FROM columns_agg ca
		LEFT JOIN constraints_agg co ON ca.table_name = co.table_name
		ORDER BY ca.table_name
	`

	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var sb strings.Builder
	for rows.Next() {
		var ddl string
		if err := rows.Scan(&ddl); err != nil {
			return "", err
		}
		sb.WriteString(ddl)
		sb.WriteString("\n\n")
	}
	return sb.String(), rows.Err()
}

func (s Stardust) Chat(tabID int64, chatModel string, aiChat []model.AIMsg) (model.AIMsg, error) {
	// Fetch Active Pool ID from Tab
	var activePoolID *string
	err := s.DB.QueryRow("SELECT active_db_id FROM tabs WHERE id = ?", tabID).Scan(&activePoolID)
	if err != nil {
		return model.AIMsg{}, errors.Wrap(err, "Tab doesn't exist")
	}
	if activePoolID == nil {
		return model.AIMsg{}, errors.New("Active pool doesn't exist in tab")
	}
	activePoolIDUUID, err := uuid.Parse(*activePoolID)
	if err != nil {
		return model.AIMsg{}, errors.Wrap(err, "Invalid active pool in tab")
	}

	pool, exists := s.PM.GetPool(activePoolIDUUID)
	if !exists {
		return model.AIMsg{}, errors.New("pool doesn't exist")
	}

	// Fetch the active session from db
	token, err := s.Auth.getToken()
	if err != nil {
		return model.AIMsg{}, err
	}

	dbmxClient := dbmx.NewClient(dbmx.Credentials{
		BaseURL:     s.Env.DBMXConfig.BaseURL,
		AccessToken: token.AccessToken,
	})

	r := dbmx.ChatReq{
		Model: chatModel,
	}

	// Inject system prompt with schema
	schemaDDL, err := s.GetSchemaContext(pool)
	if err != nil {
		return model.AIMsg{}, errors.Wrap(err, "failed to fetch schema")
	}

	systemPrompt := `You are a PostgreSQL expert assistant. You ONLY help with:
		- Writing, explaining, and optimizing SQL queries
		- Database schema design and modifications
		- PostgreSQL-specific features and best practices

		If a question is not related to SQL or database schemas, politely decline.

		Here is the database schema:

	` + schemaDDL

	r.Messages = append(r.Messages, dbmx.Message{
		Role:    "system",
		Content: systemPrompt,
	})

	for _, msg := range aiChat {
		r.Messages = append(r.Messages, dbmx.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	messageRes, err := dbmxClient.Chat(context.TODO(), r)
	if err != nil {
		return model.AIMsg{}, err
	}

	return model.AIMsg{
		ID:        messageRes.ID,
		Role:      messageRes.Role,
		Content:   messageRes.Content,
		CreatedAt: time.Now().UTC().String(),
	}, nil
}
