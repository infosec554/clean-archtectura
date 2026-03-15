package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	response "github.com/infosec554/clean-archtectura/domain/response"
	"github.com/infosec554/clean-archtectura/pkg/token"
)

type middleware struct {
	jwtManager *token.JWTManager
	logger     zerolog.Logger
}

func NewMiddleware(secret string, logger zerolog.Logger) *middleware {
	return &middleware{
		jwtManager: token.NewJWTManager(secret),
		logger:     logger,
	}
}

// JWTAuth — Bearer tokenni tekshirib, admin claims'ni context'ga yozadi
func (m *middleware) JWTAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, response.Response{
					StatusCode:  401,
					Description: "Authorization header required",
				})
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return c.JSON(http.StatusUnauthorized, response.Response{
					StatusCode:  401,
					Description: "Invalid authorization header format. Use: Bearer <token>",
				})
			}

			tokenStr := strings.TrimSpace(parts[1])
			valid, claims, err := m.jwtManager.Verify(tokenStr)
			if err != nil || !valid {
				return c.JSON(http.StatusUnauthorized, response.Response{
					StatusCode:  401,
					Description: "Invalid or expired token",
				})
			}

			if adminID, ok := claims["admin_id"].(string); ok {
				c.Set("admin_id", adminID)
			}
			if email, ok := claims["email"].(string); ok {
				c.Set("email", email)
			}
			if role, ok := claims["role"].(string); ok {
				c.Set("role", role)
			}

			return next(c)
		}
	}
}

// SuperAdminOnly — faqat superadmin o'ta oladi
func SuperAdminOnly() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if getString(c.Get("role")) != "superadmin" {
				return c.JSON(http.StatusForbidden, response.Response{
					StatusCode:  403,
					Description: "Superadmin access required",
				})
			}
			return next(c)
		}
	}
}

func getString(val any) string {
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}
