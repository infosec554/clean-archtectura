package rest

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"github.com/infosec554/clean-archtectura/domain/response"
	domain "github.com/infosec554/clean-archtectura/domain/admin"
	"github.com/infosec554/clean-archtectura/internal/rest/middleware"
)

type AdminService interface {
	Login(ctx context.Context, req *domain.LoginRequest) error
	VerifyLogin(ctx context.Context, req *domain.VerifyLoginRequest) (domain.LoginResponse, error)
	InviteAdmin(ctx context.Context, req *domain.InviteAdminRequest, invitedBy uuid.UUID) error
	AcceptInvite(ctx context.Context, req *domain.AcceptInviteRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.AdminResponse, error)
	List(ctx context.Context, limit, page int) (domain.AdminList, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, req *domain.UpdatePasswordRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type AdminHandler struct {
	service AdminService
	logger  zerolog.Logger
}

func NewAdminHandler(public *echo.Group, private *echo.Group, superadmin *echo.Group, svc AdminService, logger zerolog.Logger) {
	h := &AdminHandler{
		service: svc,
		logger:  logger.With().Str("handler", "admin").Logger(),
	}

	// Public — auth
	public.POST("/admin/login", h.Login)
	public.POST("/admin/verify-login", h.VerifyLogin)
	public.POST("/admin/accept-invite", h.AcceptInvite)

	// Private — login bo'lgan har qanday admin
	private.GET("/admin/me", h.Me)
	private.PUT("/admin/password", h.UpdatePassword)
	private.POST("/admin/logout", h.Logout)

	// Superadmin only
	superadmin.POST("/admin/invite", h.InviteAdmin)
	superadmin.GET("/admin/list", h.List)
	superadmin.DELETE("/admin/:id", h.Delete)
}

// @Summary      Admin login
// @Description  Email+parol tekshiradi, 2FA kod yuboradi
// @Tags         Admin Auth
// @Accept       json
// @Produce      json
// @Param        body body domain.LoginRequest true "Login"
// @Success      200 {object} response.Response
// @Router       /admin/login [post]
func (h *AdminHandler) Login(c echo.Context) error {
	var req domain.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Response{StatusCode: 400, Description: "Invalid payload"})
	}

	if err := h.service.Login(c.Request().Context(), &req); err != nil {
		return c.JSON(http.StatusUnauthorized, response.Response{StatusCode: 401, Description: err.Error()})
	}

	return c.JSON(http.StatusOK, response.Response{
		StatusCode:  200,
		Description: "Verification code sent to " + req.Email,
	})
}

// @Summary      Verify login (2FA)
// @Description  Emailga kelgan kodni tekshirib JWT qaytaradi
// @Tags         Admin Auth
// @Accept       json
// @Produce      json
// @Param        body body domain.VerifyLoginRequest true "Email + code"
// @Success      200 {object} response.Response{data=domain.LoginResponse}
// @Router       /admin/verify-login [post]
func (h *AdminHandler) VerifyLogin(c echo.Context) error {
	var req domain.VerifyLoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Response{StatusCode: 400, Description: "Invalid payload"})
	}

	resp, err := h.service.VerifyLogin(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, response.Response{StatusCode: 401, Description: err.Error()})
	}

	return c.JSON(http.StatusOK, response.Response{
		StatusCode:  200,
		Description: "Login successful",
		Data:        resp,
	})
}

// @Summary      Accept invite
// @Description  Taklif tokenini qabul qilib parol o'rnatadi
// @Tags         Admin Auth
// @Accept       json
// @Produce      json
// @Param        body body domain.AcceptInviteRequest true "Token + password"
// @Success      201 {object} response.Response
// @Router       /admin/accept-invite [post]
func (h *AdminHandler) AcceptInvite(c echo.Context) error {
	var req domain.AcceptInviteRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Response{StatusCode: 400, Description: "Invalid payload"})
	}

	if err := h.service.AcceptInvite(c.Request().Context(), &req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.Response{StatusCode: 422, Description: err.Error()})
	}

	return c.JSON(http.StatusCreated, response.Response{StatusCode: 201, Description: "Admin account created successfully"})
}

// @Summary      Invite admin (superadmin only)
// @Description  Yangi admin taklif qiladi
// @Tags         Admin Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body domain.InviteAdminRequest true "Email"
// @Success      200 {object} response.Response
// @Router       /admin/invite [post]
func (h *AdminHandler) InviteAdmin(c echo.Context) error {
	var req domain.InviteAdminRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Response{StatusCode: 400, Description: "Invalid payload"})
	}

	invitedBy := middleware.GetAdminID(c)

	if err := h.service.InviteAdmin(c.Request().Context(), &req, invitedBy); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Response{StatusCode: 500, Description: err.Error()})
	}

	return c.JSON(http.StatusOK, response.Response{
		StatusCode:  200,
		Description: "Invite sent to " + req.Email,
	})
}

// @Summary      Get current admin (me)
// @Tags         Admin
// @Security     BearerAuth
// @Success      200 {object} response.Response{data=domain.AdminResponse}
// @Router       /admin/me [get]
func (h *AdminHandler) Me(c echo.Context) error {
	id := middleware.GetAdminID(c)

	admin, err := h.service.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, response.Response{StatusCode: 404, Description: err.Error()})
	}

	return c.JSON(http.StatusOK, response.Response{StatusCode: 200, Description: "OK", Data: admin})
}

// @Summary      List admins (superadmin only)
// @Tags         Admin Management
// @Security     BearerAuth
// @Param        limit query int false "Limit" default(20)
// @Param        page  query int false "Page"  default(1)
// @Success      200 {object} response.Response{data=domain.AdminList}
// @Router       /admin/list [get]
func (h *AdminHandler) List(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}

	list, err := h.service.List(c.Request().Context(), limit, page)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Response{StatusCode: 500, Description: err.Error()})
	}

	return c.JSON(http.StatusOK, response.Response{StatusCode: 200, Description: "OK", Data: list})
}

// @Summary      Update password
// @Tags         Admin
// @Security     BearerAuth
// @Param        body body domain.UpdatePasswordRequest true "Passwords"
// @Success      200 {object} response.Response
// @Router       /admin/password [put]
func (h *AdminHandler) UpdatePassword(c echo.Context) error {
	var req domain.UpdatePasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Response{StatusCode: 400, Description: "Invalid payload"})
	}

	id := middleware.GetAdminID(c)
	if err := h.service.UpdatePassword(c.Request().Context(), id, &req); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Response{StatusCode: 500, Description: err.Error()})
	}

	return c.JSON(http.StatusOK, response.Response{StatusCode: 200, Description: "Password updated"})
}

// @Summary      Delete admin (superadmin only)
// @Tags         Admin Management
// @Security     BearerAuth
// @Param        id path string true "Admin ID"
// @Success      200 {object} response.Response
// @Router       /admin/{id} [delete]
func (h *AdminHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Response{StatusCode: 400, Description: "Invalid ID"})
	}

	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		code := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			code = http.StatusNotFound
		}
		return c.JSON(code, response.Response{StatusCode: code, Description: err.Error()})
	}

	return c.JSON(http.StatusOK, response.Response{StatusCode: 200, Description: "Admin deleted"})
}

// @Summary      Logout
// @Tags         Admin Auth
// @Security     BearerAuth
// @Success      200 {object} response.Response
// @Router       /admin/logout [post]
func (h *AdminHandler) Logout(c echo.Context) error {
	return c.JSON(http.StatusOK, response.Response{StatusCode: 200, Description: "Logged out successfully"})
}
