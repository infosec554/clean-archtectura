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

func GetUserType(c echo.Context) string {
	return getString(c.Get("user_type"))
}

func GetEmail(c echo.Context) string {
	return getString(c.Get("email"))
}

func GetPINFL(c echo.Context) string {
	return getString(c.Get("pinfl"))
}

func GetCompanyID(c echo.Context) uuid.UUID {
	id := getString(c.Get("company_id"))
	parsed, _ := uuid.Parse(id)
	return parsed
}

func HasCompany(c echo.Context) bool {
	return getString(c.Get("company_id")) != ""
}

func GetFirstName(c echo.Context) string {
	return getString(c.Get("first_name"))
}

func GetLastName(c echo.Context) string {
	return getString(c.Get("last_name"))
}

func IsStudent(c echo.Context) bool {
	return GetUserType(c) == "student"
}

func IsUser(c echo.Context) bool {
	return GetUserType(c) == "user"
}

func IsCompany(c echo.Context) bool {
	return GetUserType(c) == "company"
}

func IsUniversity(c echo.Context) bool {
	return GetUserType(c) == "university"
}
