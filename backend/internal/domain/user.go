package domain

import (
	"strings"
	"time"
)

type Role string

const (
	RoleRequester Role = "requester"
	RoleAgent     Role = "agent"
	RoleAdmin     Role = "admin"
)

func (r Role) Valid() bool {
	switch r {
	case RoleRequester, RoleAgent, RoleAdmin:
		return true
	default:
		return false
	}
}

func (r Role) IsStaff() bool {
	return r == RoleAgent || r == RoleAdmin
}

func ParseRole(raw string) (Role, bool) {
	r := Role(strings.ToLower(strings.TrimSpace(raw)))
	return r, r.Valid()
}

type User struct {
	ID           int64
	Name         string
	Email        string
	PasswordHash string
	Role         Role
	CreatedAt    time.Time
}

// UserSummary is the safe public representation of a user.
type UserSummary struct {
	ID    int64
	Name  string
	Email string
	Role  Role
}
