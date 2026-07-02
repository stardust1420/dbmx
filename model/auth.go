package model

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	Fullname string    `json:"fullname"`

	CustomerID    string `json:"customer_id"`
	UseDefaultKey bool   `json:"use_default_key"`
}
