package domain

import "github.com/google/uuid"

type UserID uuid.UUID

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

type User struct {
	ID    UserID
	Email Email
	Role  UserRole
}

type Token string
