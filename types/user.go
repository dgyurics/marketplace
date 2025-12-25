package types

import (
	"time"
)

type PendingUser struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CodeHash  string    `json:"-"`
	Used      bool      `json:"used"`
	ExpiresAt time.Time `json:"expires_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	Password      string    `json:"-"`
	PasswordHash  string    `json:"-"`
	Role          Role      `json:"role"`
	RequiresSetup bool      `json:"requires_setup"`
	UpdatedAt     time.Time `json:"updated_at"`
	CreatedAt     time.Time `json:"created_at"`
}

type Role string

const (
	RoleGuest  Role = "guest"
	RoleUser   Role = "user"
	RoleMember Role = "member"
	RoleStaff  Role = "staff"
	RoleAdmin  Role = "admin"
)

var hierarchy = map[Role]int{
	RoleGuest:  0,
	RoleUser:   1,
	RoleMember: 2,
	RoleStaff:  3,
	RoleAdmin:  4,
}

// HasMinimumRole checks if the user has a role equal to
// or higher than the specified role in the hierarchy.
func (u *User) HasMinimumRole(role Role) bool {
	return hierarchy[Role(u.Role)] >= hierarchy[role]
}
