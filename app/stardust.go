package app

import (
	"context"
	"database/sql"
	"dbmx/config/env"

	"github.com/pkg/errors"
	"github.com/stardust1420/dbmx-go"
)

type Stardust struct {
	DB   *sql.DB
	Env  *env.Env
	Auth *Auth
}

func NewStardust(db *sql.DB, env *env.Env, auth *Auth) *Stardust {
	return &Stardust{
		DB:   db,
		Env:  env,
		Auth: auth,
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

func (s *Stardust) ListUserProviders() ([]dbmx.UserProvider, error) {
	// Fetch the active session from db
	token, err := s.Auth.getToken()
	if err != nil {
		return nil, err
	}

	dbmxClient := dbmx.NewClient(dbmx.Credentials{
		BaseURL:     s.Env.DBMXConfig.BaseURL,
		AccessToken: token.AccessToken,
	})

	userProviders, err := dbmxClient.ListUserProviders(context.TODO())
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
