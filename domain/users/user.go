package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserStatus string

func (s UserStatus) String() string {
	return string(s)
}

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusPending  UserStatus = "pending"
)

// User represents a system user
type User struct {
	ID         uuid.UUID `json:"id" db:"id"`
	PINFL      string    `json:"pinfl" db:"pinfl"`
	Passport   *string   `json:"passport" db:"passport"`
	FirstName  string    `json:"first_name" db:"first_name"`
	LastName   string    `json:"last_name" db:"last_name"`
	MiddleName *string   `json:"middle_name,omitempty" db:"middle_name"`
	Email      *string   `json:"email,omitempty" db:"email"`
	Phone      *string   `json:"phone,omitempty" db:"phone"`
	Password   *string   `json:"-" db:"password"`
	Type       string    `json:"-" db:"company_user_type"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// CreateUser request for creating a new user
type CreateUser struct {
	PINFL      string `json:"pinfl" validate:"required,len=14,numeric"`
	Passport   string `json:"passport" validate:"required,min=2,max=50"`
	FirstName  string `json:"first_name" validate:"required,min=2,max=100"`
	LastName   string `json:"last_name" validate:"required,min=2,max=100"`
	MiddleName string `json:"middle_name,omitempty"`
	Email      string `json:"email,omitempty" validate:"omitempty,email"`
	Phone      string `json:"phone,omitempty"`
	Password   string `json:"password,omitempty" validate:"omitempty,min=6,max=50"`
}

// UpdateUser request for updating a user
type UpdateUser struct {
	ID         uuid.UUID `json:"-" validate:"required"`
	PINFL      string    `json:"pinfl,omitempty" validate:"omitempty,len=14,numeric"`
	Passport   string    `json:"passport,omitempty"`
	FirstName  string    `json:"first_name,omitempty"`
	LastName   string    `json:"last_name,omitempty"`
	MiddleName string    `json:"middle_name,omitempty"`
	Email      string    `json:"email,omitempty" validate:"omitempty,email"`
	Phone      string    `json:"phone,omitempty"`
	Password   string    `json:"password,omitempty" validate:"omitempty,min=6,max=50"`
}

// UserResponse for returning user data
type UserResponse struct {
	ID         uuid.UUID `json:"id"`
	PINFL      string    `json:"pinfl"`
	Passport   string    `json:"passport"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	MiddleName string    `json:"middle_name,omitempty"`
	FullName   string    `json:"full_name"`
	Email      string    `json:"email,omitempty"`
	Phone      string    `json:"phone,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// UserList for paginated user listing
type UserList struct {
	List []UserResponse `json:"list"`
	Meta Meta           `json:"meta"`
}

type Meta struct {
	Total     int `json:"total"`
	Page      int `json:"page"`
	PageSize  int `json:"page_size"`
	PageCount int `json:"page_count"`
}

// CompanyUser represents a user's association with a company and role
type CompanyUser struct {
	CompanyID uuid.UUID `json:"company_id" db:"company_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	RoleID    uuid.UUID `json:"role_id" db:"role_id"`
	Status    string    `json:"status" db:"status"`
}

// CreateCompanyUser request for assigning a user to a company with a role
type CreateCompanyUser struct {
	CompanyID string `json:"company_id" validate:"required,uuid"`
	UserID    string `json:"user_id" validate:"required,uuid"`
	RoleID    string `json:"role_id" validate:"required,uuid"`
	Status    string `json:"-" validate:"required,oneof=active inactive pending"`
	Type      string `json:"type"`
}

// UpdateCompanyUser request for updating a company user
type UpdateCompanyUser struct {
	CompanyID uuid.UUID `json:"-"`
	UserID    uuid.UUID `json:"-"`
	RoleID    string    `json:"role_id,omitempty" validate:"omitempty,uuid"`
	Status    string    `json:"status,omitempty" validate:"omitempty,oneof=active inactive pending"`
}

// CompanyUserResponse for returning company user data
type CompanyUserResponse struct {
	CompanyID   uuid.UUID    `json:"company_id"`
	CompanyName string       `json:"company_name,omitempty"`
	User        UserResponse `json:"user"`
	Role        RoleResponse `json:"role"`
	Status      string       `json:"status"`
}

// CompanyUserList for paginated company user listing
type CompanyUserList struct {
	List []CompanyUserResponse `json:"list"`
	Meta Meta                  `json:"meta"`
}

// MeCompanyResponse represents a company with role and permissions for the Me endpoint
type MeCompanyResponse struct {
	CompanyID   uuid.UUID `json:"company_id"`
	CompanyName string    `json:"company_name"`
	Role        RoleInfo  `json:"role"`
	Permissions []string  `json:"permissions"`
	Status      string    `json:"status"`
	UserType    string    `json:"user_type"`
}

// RoleInfo represents basic role information
type RoleInfo struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
}

// MeResponse for the /me endpoint
type MeResponse struct {
	User    UserResponse      `json:"user"`
	Company MeCompanyResponse `json:"company"`
}

type AddUserRequest struct {
	PINFL    string `json:"pinfl" validate:"required,len=14"`
	Passport string `json:"passport" validate:"required"`
	RoleID   string `json:"role_id" validate:"required,uuid"`
}
