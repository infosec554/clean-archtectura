package domain

import (
	"time"

	"github.com/google/uuid"
)

type AdminRole string

const (
	RoleSuperAdmin AdminRole = "superadmin"
	RoleAdmin      AdminRole = "admin"
)

// Admin — admins jadvali bilan mos
type Admin struct {
	ID        uuid.UUID  `json:"id"         db:"id"`
	Email     string     `json:"email"      db:"email"`
	Password  string     `json:"-"          db:"password"`
	Role      AdminRole  `json:"role"       db:"role"`
	InvitedBy *uuid.UUID `json:"invited_by" db:"invited_by"`
	IsActive  bool       `json:"is_active"  db:"is_active"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// AdminResponse — frontendga qaytariladigan ma'lumot (password yo'q)
type AdminResponse struct {
	ID        uuid.UUID  `json:"id"`
	Email     string     `json:"email"`
	Role      AdminRole  `json:"role"`
	InvitedBy *uuid.UUID `json:"invited_by,omitempty"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
}

// LoginRequest — email + password
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse — muvaffaqiyatli login javob
type LoginResponse struct {
	Admin        AdminResponse `json:"admin"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
}

// VerifyLoginRequest — 2FA: login paytida kelgan kod
type VerifyLoginRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code"  validate:"required,len=6"`
}

// InviteAdminRequest — superadmin yangi admin taklif qiladi
type InviteAdminRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// AcceptInviteRequest — yangi admin taklif tokenini qabul qilib parol o'rnatadi
type AcceptInviteRequest struct {
	Token    string `json:"token"    validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

// UpdatePasswordRequest — parolni yangilash
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

// Meta — pagination
type Meta struct {
	Total     int `json:"total"`
	Page      int `json:"page"`
	PageSize  int `json:"page_size"`
	PageCount int `json:"page_count"`
}

// AdminList — sahifalangan adminlar ro'yxati
type AdminList struct {
	List []AdminResponse `json:"list"`
	Meta Meta            `json:"meta"`
}
