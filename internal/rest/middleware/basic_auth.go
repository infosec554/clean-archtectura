package middleware

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	response "github.com/infosec554/clean-archtectura/domain/response"
)

// BasicAuth validates basic auth credentials against provided username and password
func BasicAuth(username, password string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, response.Response{
					StatusCode:  401,
					Description: "Authorization header required",
				})
			}

			if !strings.HasPrefix(authHeader, "Basic ") {
				return c.JSON(http.StatusUnauthorized, response.Response{
					StatusCode:  401,
					Description: "Invalid authorization header format. Use Basic auth",
				})
			}

			encoded := strings.TrimPrefix(authHeader, "Basic ")
			decoded, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, response.Response{
					StatusCode:  401,
					Description: "Invalid base64 encoding",
				})
			}

			credentials := strings.SplitN(string(decoded), ":", 2)
			if len(credentials) != 2 {
				return c.JSON(http.StatusUnauthorized, response.Response{
					StatusCode:  401,
					Description: "Invalid credentials format",
				})
			}

			if credentials[0] != username || credentials[1] != password {
				return c.JSON(http.StatusUnauthorized, response.Response{
					StatusCode:  401,
					Description: "Invalid credentials",
				})
			}

			return next(c)
		}
	}
}
