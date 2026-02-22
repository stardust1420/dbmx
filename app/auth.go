package app

import "database/sql"

type Auth struct {
	DB           *sql.DB
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func NewAuth(db *sql.DB) *Auth {
	return &Auth{
		DB: db,
	}
}

func (a *Auth) Login() {

}

func (a *Auth) Logout() {

}
