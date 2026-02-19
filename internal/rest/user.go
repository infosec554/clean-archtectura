package rest

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"github.com/infosec554/clean-archtectura/config"
	"github.com/infosec554/clean-archtectura/domain/response"
	domain "github.com/infosec554/clean-archtectura/domain/users"
	"github.com/infosec554/clean-archtectura/pkg/cache"
)

type UserService interface {
	Register(ctx context.Context, req *domain.CreateUser) (string, error)
	Login(ctx context.Context, req *domain.LoginRequest) (domain.LoginResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.UserResponse, error)
	Update(ctx context.Context, req *domain.UpdateUser) (string, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, req *domain.UpdatePasswordRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type UserHandler struct {
	cache   cache.ICache
	config  config.Config
	service UserService
	logger  zerolog.Logger
}

func NewUserHandler(public *echo.Group, private *echo.Group, svc UserService, cfg config.Config, c cache.ICache, logger zerolog.Logger) {
	h := &UserHandler{
		service: svc,
		cache:   c,
		config:  cfg,
		logger:  logger.With().Str("handler", "user").Logger(),
	}

	// Public routes
	public.POST("/register", h.Register)
	public.POST("/login", h.Login)

	// Private routes
	private.POST("/logout", h.Logout)
	private.GET("/users/:id", h.GetByID)
	private.PUT("/users/:id", h.Update)
	private.PUT("/users/:id/password", h.UpdatePassword)
	private.DELETE("/users/:id", h.Delete)
}

// @Summary      Get user by ID
// @Description  Returns user details by UUID
// @Tags         Users
// @Produce      json
// @Param        id path string true "User ID"
// @Security     BearerAuth
// @Success      200 {object} response.Response "User retrieved"
// @Failure      400 {object} response.Response "Invalid user ID"
// @Failure      404 {object} response.Response "User not found"
// @Failure      500 {object} response.Response "Internal server error"
// @Router       /users/{id} [get]
func (h *UserHandler) GetByID(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Response{
			StatusCode:  400,
			Description: "Invalid user ID",
		})
	}

	user, err := h.service.GetByID(c.Request().Context(), id)
	if err != nil {
		code := http.StatusInternalServerError
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			code = http.StatusNotFound
		}
		return c.JSON(code, response.Response{
			StatusCode:  code,
			Description: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response.Response{
		StatusCode:  200,
		Description: "User retrieved",
		Data:        user,
	})
}

// @Summary      Update user
// @Description  Updates existing user by ID
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID"
// @Param        user body domain.UpdateUser true "Updated user info"
// @Security     BearerAuth
// @Success      200 {object} response.Response "User updated successfully"
// @Failure      400 {object} response.Response "Invalid request"
// @Failure      404 {object} response.Response "User not found"
// @Failure      422 {object} response.Response "Validation failed"
// @Failure      500 {object} response.Response "Internal server error"
// @Router       /users/{id} [put]
func (h *UserHandler) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Response{
			StatusCode:  400,
			Description: "Invalid user ID",
		})
	}

	var req domain.UpdateUser
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Response{
			StatusCode:  400,
			Description: "Invalid payload",
			Data:        err.Error(),
		})
	}
	req.ID = id

	if _, err := h.service.Update(c.Request().Context(), &req); err != nil {
		code := http.StatusInternalServerError
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			code = http.StatusNotFound
		}
		return c.JSON(code, response.Response{
			StatusCode:  code,
			Description: "Failed to update user",
			Data:        err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response.Response{
		StatusCode:  200,
		Description: "User updated successfully",
		Data: map[string]string{
			"id": req.ID.String(),
		},
	})
}

// @Summary      Delete user
// @Description  Deletes a user by ID
// @Tags         Users
// @Produce      json
// @Param        id path string true "User ID"
// @Security     BearerAuth
// @Success      200 {object} response.Response "User deleted successfully"
// @Failure      400 {object} response.Response "Invalid user ID"
// @Failure      404 {object} response.Response "User not found"
// @Failure      500 {object} response.Response "Internal server error"
// @Router       /users/{id} [delete]
func (h *UserHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Response{
			StatusCode:  400,
			Description: "Invalid user ID",
		})
	}

	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		code := http.StatusInternalServerError
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			code = http.StatusNotFound
		}
		return c.JSON(code, response.Response{
			StatusCode:  code,
			Description: "Failed to delete user",
			Data:        err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response.Response{
		StatusCode:  200,
		Description: "User deleted successfully",
	})
}

// @Summary      Register user
// @Description  Creates a new user account
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        user body domain.CreateUser true "Registration info"
// @Success      201 {object} response.Response "User registered"
// @Router       /register [post]
func (h *UserHandler) Register(c echo.Context) error {
	var req domain.CreateUser
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Response{
			StatusCode:  400,
			Description: "Invalid payload",
		})
	}

	id, err := h.service.Register(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Response{
			StatusCode:  500,
			Description: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, response.Response{
		StatusCode:  201,
		Description: "User registered successfully",
		Data:        map[string]string{"id": id},
	})
}

// @Summary      Login
// @Description  Authenticates user and returns JWT tokens
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        login body domain.LoginRequest true "Login credentials"
// @Success      200 {object} response.Response{data=domain.LoginResponse} "Login success"
// @Router       /login [post]
func (h *UserHandler) Login(c echo.Context) error {
	var req domain.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Response{
			StatusCode:  400,
			Description: "Invalid payload",
		})
	}

	resp, err := h.service.Login(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, response.Response{
			StatusCode:  401,
			Description: "Invalid credentials",
		})
	}

	return c.JSON(http.StatusOK, response.Response{
		StatusCode:  200,
		Description: "Login successful",
		Data:        resp,
	})
}

// @Summary      Logout
// @Description  Invalidates the current session (client must remove the token)
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.Response "Logged out"
// @Router       /logout [post]
func (h *UserHandler) Logout(c echo.Context) error {
	return c.JSON(http.StatusOK, response.Response{
		StatusCode:  200,
		Description: "Logged out successfully",
	})
}

// @Summary      Update password
// @Description  Updates authenticated user's password
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID"
// @Param        update body domain.UpdatePasswordRequest true "Password update info"
// @Security     BearerAuth
// @Success      200 {object} response.Response "Password updated"
// @Router       /users/{id}/password [put]
func (h *UserHandler) UpdatePassword(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Response{
			StatusCode:  400,
			Description: "Invalid user ID",
		})
	}

	var req domain.UpdatePasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Response{
			StatusCode:  400,
			Description: "Invalid payload",
		})
	}

	if err := h.service.UpdatePassword(c.Request().Context(), id, &req); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Response{
			StatusCode:  500,
			Description: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response.Response{
		StatusCode:  200,
		Description: "Password updated successfully",
	})
}
