package app

import (
	"database/sql"
	"dbmx/config/env"
	"dbmx/model"
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"
	"github.com/supabase-community/auth-go"
	"github.com/supabase-community/auth-go/types"
)

type Auth struct {
	DB *sql.DB

	// Client to interact with Supabase Auth API
	SupabaseClient auth.Client

	// This client will be used to make requests to the Supabase Auth API with the logged in user's access token
	SupabaseAuthedClient auth.Client

	IsLoggedIn bool

	// Logged in user's details
	User *types.User
}

func InitAuth(db *sql.DB, supabaseConfig env.SupabaseConfig) *Auth {
	// Init Supabase client
	client := auth.New(
		supabaseConfig.ProjectID,
		supabaseConfig.AnonKey,
	)

	a := &Auth{
		DB:             db,
		SupabaseClient: client,
		IsLoggedIn:     false,
	}

	a.getLatestSession()

	return a
}

// Fetch the latest session from db and set the logged in user
func (a *Auth) getLatestSession() {
	var err error

	// Fetch the latest session from db
	var token types.TokenResponse
	err = a.DB.QueryRow("SELECT access_token, refresh_token, expires_at FROM active_session").Scan(&token.AccessToken, &token.RefreshToken, &token.ExpiresAt)
	if err != nil {
		log.Println("Error getting latest session:", err)
		return
	}

	// Check for expired access token token
	if time.Now().UTC().Unix() > token.ExpiresAt {
		fmt.Println("token expired, refreshing...")
		// Token expired, try to refresh
		// TODO: handle multi retries
		token, err = a.refreshSession(token.RefreshToken)
		if err != nil {
			log.Println("Error refreshing token:", err)
			return
		}
	}

	// Set logged in user's params
	a.SupabaseAuthedClient = a.SupabaseClient.WithToken(token.AccessToken)
	a.IsLoggedIn = true

	// Fetch user from supabase
	userRes, err := a.SupabaseAuthedClient.GetUser()
	if err != nil {
		// TODO handle expired token
		log.Println("Error getting user:", err)
		return
	}

	// Store logged in user in memory
	a.User = &userRes.User
}

func (a *Auth) SignUp(fullname, email, password, confirmPassword string) (bool, error) {
	if password != confirmPassword {
		return false, errors.New("passwords do not match")
	}

	token, err := a.SupabaseClient.Signup(types.SignupRequest{
		Email:    email,
		Password: password,
		Data: map[string]interface{}{
			"fullname": fullname,
		},
	})
	if err != nil {
		return false, err
	}

	// Save session in db
	if err := a.saveSession(token.Session); err != nil {
		return false, err
	}

	// Create a new client with the user's access token
	authedClient := a.SupabaseClient.WithToken(
		token.AccessToken,
	)

	// Store the authed client
	a.SupabaseAuthedClient = authedClient
	a.IsLoggedIn = true

	// Get User
	userRes, err := a.SupabaseAuthedClient.GetUser()
	if err != nil {
		// TODO handle expired token
		log.Println("Error getting user:", err)
		return false, err
	}

	// Store logged in user in memory
	a.User = &userRes.User

	return true, nil
}

func (a *Auth) Login(email, password string) (bool, error) {
	// Log in a user (get access and refresh tokens)
	token, err := a.SupabaseClient.Token(types.TokenRequest{
		GrantType: "password",
		Email:     email,
		Password:  password,
	})
	if err != nil {
		return false, err
	}

	// Save session in db
	if err := a.saveSession(token.Session); err != nil {
		return false, err
	}

	// Create a new client with the user's access token
	authedClient := a.SupabaseClient.WithToken(
		token.AccessToken,
	)

	// Store the authed client
	a.SupabaseAuthedClient = authedClient
	a.IsLoggedIn = true

	// Get User
	userRes, err := a.SupabaseAuthedClient.GetUser()
	if err != nil {
		// TODO handle expired token
		log.Println("Error getting user:", err)
		return false, err
	}

	// Store logged in user in memory
	a.User = &userRes.User

	return true, nil
}

func (a *Auth) saveSession(token types.Session) error {
	// Delete the active session from db
	_, err := a.DB.Exec("DELETE FROM active_session")
	if err != nil {
		return err
	}

	// Insert the new session
	_, err = a.DB.Exec("INSERT INTO active_session (access_token, refresh_token, expires_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", token.AccessToken, token.RefreshToken, token.ExpiresAt, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		return err
	}
	return nil
}

// Get Logged In User from memory
func (a *Auth) GetLoggedInUser() (model.User, error) {
	if !a.IsLoggedIn {
		return model.User{}, errors.New("not logged in")
	}
	if a.User == nil {
		return model.User{}, errors.New("user not found")
	}
	return model.User{
		ID:       a.User.ID,
		Email:    a.User.Email,
		Fullname: a.User.UserMetadata["fullname"].(string),
	}, nil
}

func (a *Auth) refreshSession(refreshToken string) (types.TokenResponse, error) {
	token, err := a.SupabaseClient.RefreshToken(refreshToken)
	if err != nil {
		return types.TokenResponse{}, err
	}

	// Update session in db
	if err := a.saveSession(token.Session); err != nil {
		return types.TokenResponse{}, err
	}

	// Create a new client with the user's access token
	authedClient := a.SupabaseClient.WithToken(
		token.AccessToken,
	)

	// Store the authed client
	a.SupabaseAuthedClient = authedClient
	a.IsLoggedIn = true

	return *token, nil
}

func (a *Auth) Logout() error {
	err := a.SupabaseAuthedClient.Logout()
	if err != nil {
		return err
	}
	// Delete the active session from db
	_, err = a.DB.Exec("DELETE FROM active_session")
	if err != nil {
		return err
	}
	a.SupabaseAuthedClient = nil
	a.IsLoggedIn = false
	a.User = nil
	return nil
}
