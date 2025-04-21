package models

import "github.com/google/uuid"

const (
	RoleClient   = "client"
	RoleModerator = "moderator"
	RoleWorker   = "worker"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash []byte    `json:"-"`
	Role         string    `json:"role"`
}