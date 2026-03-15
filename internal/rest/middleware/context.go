package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func GetAdminID(c echo.Context) uuid.UUID {
	parsed, _ := uuid.Parse(getString(c.Get("admin_id")))
	return parsed
}

func GetEmail(c echo.Context) string {
	return getString(c.Get("email"))
}

func GetRole(c echo.Context) string {
	return getString(c.Get("role"))
}

func IsSuperAdmin(c echo.Context) bool {
	return GetRole(c) == "superadmin"
}
