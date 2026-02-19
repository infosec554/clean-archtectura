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
	"github.com/infosec554/clean-archtectura/pkg/cache"

	domain "github.com/infosec554/clean-archtectura/domain/users"
)

type UserService interface {
	GetByID(ctx context.Context, id uuid.UUID) (domain.UserResponse, error)
	Update(ctx context.Context, req *domain.UpdateUser) (string, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type UserHandler struct {
	cache   cache.ICache
	config  config.Config
	service UserService
	logger  zerolog.Logger
}

func NewUserHandler(g *echo.Group, svc UserService, cfg config.Config, c cache.ICache, logger zerolog.Logger) {
	h := &UserHandler{
		service: svc,
		cache:   c,
		config:  cfg,
		logger:  logger.With().Str("handler", "user").Logger(),
	}

	g.GET("/users/:id", h.GetByID)
	g.PUT("/users/:id", h.Update)
	g.DELETE("/users/:id", h.Delete)
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

	if ok, err := isValid(&req); !ok {
		return c.JSON(http.StatusUnprocessableEntity, response.Response{
			StatusCode:  422,
			Description: "Validation failed",
			Data:        err.Error(),
		})
	}

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
