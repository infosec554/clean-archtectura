package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a system user
type User struct {
	ID            uuid.UUID `json:"id" db:"id"`
	FirstName     string    `json:"first_name" db:"first_name"`
	LastName      string    `json:"last_name" db:"last_name"`
	Email         *string   `json:"email,omitempty" db:"email"`
	Password      *string   `json:"-" db:"password"`
	EmailVerified bool      `json:"email_verified" db:"email_verified"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
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
	ID            uuid.UUID `json:"id"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	Email         string    `json:"email,omitempty"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// VerifyEmailRequest — email + code
type VerifyEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=6"`
}

// ResendCodeRequest — faqat email
type ResendCodeRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// LoginRequest ...
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse ...
type LoginResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

// ResetPasswordRequest ...
type ResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// UpdatePasswordRequest ...
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
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
