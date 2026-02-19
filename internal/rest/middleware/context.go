package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func GetUserID(c echo.Context) uuid.UUID {
	id := getString(c.Get("user_id"))
	parsed, _ := uuid.Parse(id)
	return parsed
}

func GetEmail(c echo.Context) string {
	return getString(c.Get("email"))
}

func GetFirstName(c echo.Context) string {
	return getString(c.Get("first_name"))
}

func GetLastName(c echo.Context) string {
	return getString(c.Get("last_name"))
}
