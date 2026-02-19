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

// JWTAuth — Validate bearer token and attach claims to request context
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
					Description: "Invalid authorization header format",
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

			// Extract User ID and Email
			if userID, ok := claims["user_id"].(string); ok {
				c.Set("user_id", userID)
			}
			if email, ok := claims["email"].(string); ok {
				c.Set("email", email)
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
