package model

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	Fullname string    `json:"fullname"`
}
