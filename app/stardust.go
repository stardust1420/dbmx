package app

import (
	"context"
	"database/sql"
	"dbmx/config/env"

	"github.com/pkg/errors"
	"github.com/stardust1420/dbmx-go"
	"github.com/supabase-community/auth-go/types"
)

type Stardust struct {
	DB  *sql.DB
	Env *env.Env
}

func NewStardust(db *sql.DB, env *env.Env) *Stardust {
	return &Stardust{
		DB:  db,
		Env: env,
	}
}

func (s *Stardust) EnableStardustAI() (dbmx.Customer, error) {
	// Fetch the active session from db
	var token types.TokenResponse
	err := s.DB.QueryRow("SELECT access_token, refresh_token, expires_at FROM active_session").Scan(&token.AccessToken, &token.RefreshToken, &token.ExpiresAt)
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
	var token types.TokenResponse
	err := s.DB.QueryRow("SELECT access_token, refresh_token, expires_at FROM active_session").Scan(&token.AccessToken, &token.RefreshToken, &token.ExpiresAt)
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
	var token types.TokenResponse
	err := s.DB.QueryRow("SELECT access_token, refresh_token, expires_at FROM active_session").Scan(&token.AccessToken, &token.RefreshToken, &token.ExpiresAt)
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
