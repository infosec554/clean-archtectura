package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a system user
type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	Email     *string   `json:"email,omitempty" db:"email"`
	Password  *string   `json:"-" db:"password"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateUser request for creating a new user
type CreateUser struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string `json:"last_name" validate:"required,min=2,max=100"`
	Email     string `json:"email,omitempty" validate:"omitempty,email"`
	Password  string `json:"password,omitempty" validate:"omitempty,min=6,max=50"`
}

// UpdateUser request for updating a user
type UpdateUser struct {
	ID        uuid.UUID `json:"-" validate:"required"`
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	Email     string    `json:"email,omitempty" validate:"omitempty,email"`
	Password  string    `json:"password,omitempty" validate:"omitempty,min=6,max=50"`
}

// UserResponse for returning user data
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Password  string    `json:"password,omitempty" validate:"omitempty,min=6,max=50"`
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
