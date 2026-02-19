package middleware

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	echo_middleware "github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"

	response "github.com/infosec554/clean-archtectura/domain/response"
	"github.com/infosec554/clean-archtectura/pkg/token"
)

type middleware struct {
	jwtManager          *token.JWTManager
	logger              zerolog.Logger
	publicRoutesSkipper echo_middleware.Skipper
}

func NewMiddleware(secret string, publicRoutes map[string]bool, logger zerolog.Logger) *middleware {
	publicRoutesSkipper := func(c echo.Context) bool {
		return publicRoutes[c.Path()]
	}

	return &middleware{
		jwtManager:          token.NewJWTManager(secret),
		publicRoutesSkipper: publicRoutesSkipper,
		logger:              logger,
	}
}

// JWTAuth — Validate bearer token and attach claims to request context
func (m *middleware) JWTAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			if m.publicRoutesSkipper(c) {
				return next(c)
			}

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
					Description: "Invalid authorization header format. Use Bearer <token>",
				})
			}

			tokenStr := strings.TrimSpace(parts[1])
			if tokenStr == "" {
				return c.JSON(http.StatusUnauthorized, response.Response{
					StatusCode:  401,
					Description: "Token missing after Bearer",
				})
			}

			valid, claims, err := m.jwtManager.Verify(tokenStr)
			if err != nil || !valid {
				return c.JSON(http.StatusUnauthorized, response.Response{
					StatusCode:  401,
					Description: "Invalid or expired token",
				})
			}

			// Extract USER ID
			if userID, ok := claims["user_id"].(string); ok {
				if _, err := uuid.Parse(userID); err != nil {
					return c.JSON(http.StatusUnauthorized, response.Response{
						StatusCode:  401,
						Description: "Invalid or expired token",
					})
				}
				c.Set("user_id", userID)
				c.Set("student_id", userID)
			}

			// Extract ROLE (user_type)
			if userType, ok := claims["user_type"].(string); ok {
				c.Set("user_type", userType) // "student" | "company" | "university"
			}

			// Extract Email
			if email, ok := claims["email"].(string); ok {
				c.Set("email", email)
			}

			// Extract Names
			if firstName, ok := claims["first_name"].(string); ok {
				c.Set("first_name", firstName)
			}
			if lastName, ok := claims["last_name"].(string); ok {
				c.Set("last_name", lastName)
			}

			// Extract PINFL
			if pinfl, ok := claims["pinfl"].(string); ok {
				c.Set("pinfl", pinfl)
			}

			// Extract company_id (FOR COMPANY **AND** UNIVERSITY)
			if companyID, ok := claims["company_id"].(string); ok {
				if _, err := uuid.Parse(companyID); err == nil {
					c.Set("company_id", companyID)
				}
			}

			// LOG
			m.logger.Debug().
				Str("user_id", getString(c.Get("user_id"))).
				Str("user_type", getString(c.Get("user_type"))).
				Str("company_id", getString(c.Get("company_id"))).
				Msg("JWT authentication success")

			return next(c)
		}
	}
}

// RequireRole — middleware to allow only specific roles
func (m *middleware) RequireRole(allowedRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			role := getString(c.Get("user_type")) // IMPORTANT!

			if role == "" {
				return c.JSON(http.StatusForbidden, response.Response{
					StatusCode:  403,
					Description: "Access denied: user role missing",
				})
			}

			for _, allowed := range allowedRoles {
				if role == allowed {
					return next(c)
				}
			}

			return c.JSON(http.StatusForbidden, response.Response{
				StatusCode:  403,
				Description: "Access denied: insufficient permissions",
			})
		}
	}
}

// Utility
func getString(val any) string {
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}
