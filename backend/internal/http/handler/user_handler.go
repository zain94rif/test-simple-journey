package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"helpdesk/backend/internal/domain"
	"helpdesk/backend/internal/http/middleware"
	"helpdesk/backend/internal/http/response"
	"helpdesk/backend/internal/service"
)

type UserHandler struct {
	users *service.UserService
}

func NewUserHandler(users *service.UserService) *UserHandler {
	return &UserHandler{users: users}
}

// List godoc
// @Summary      Daftar pengguna (admin)
// @Description  Khusus admin. Mendukung pencarian nama/email, filter peran, dan pagination.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        page      query     int     false  "Halaman (default 1)"
// @Param        per_page  query     int     false  "Jumlah per halaman (1-100, default 20)"
// @Param        role      query     string  false  "Filter peran"  Enums(requester, agent, admin)
// @Param        search    query     string  false  "Cari pada nama atau email"
// @Success      200  {object}  UserListEnvelope
// @Failure      400  {object}  response.ErrorEnvelope
// @Failure      401  {object}  response.ErrorEnvelope
// @Failure      403  {object}  response.ErrorEnvelope
// @Router       /users [get]
func (h *UserHandler) List(c *gin.Context) {
	page, appErr := parsePositiveIntQuery(c, "page", 1)
	if appErr != nil {
		response.ErrorFrom(c, appErr)
		return
	}
	perPage, appErr := parsePositiveIntQuery(c, "per_page", 20)
	if appErr != nil {
		response.ErrorFrom(c, appErr)
		return
	}
	filter := domain.UserFilter{
		Page:    page,
		PerPage: perPage,
		Search:  strings.TrimSpace(c.Query("search")),
	}
	if raw := strings.TrimSpace(c.Query("role")); raw != "" {
		role, ok := domain.ParseRole(raw)
		if !ok {
			response.ErrorFrom(c, domain.Validation(map[string]string{"role": "Peran tidak valid."}))
			return
		}
		filter.Role = &role
	}

	users, total, err := h.users.List(c.Request.Context(), middleware.CurrentUser(c), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	items := make([]UserResponse, 0, len(users))
	for _, user := range users {
		items = append(items, toUserResponse(user))
	}
	response.List(c, items, response.MetaFrom(page, perPage, total))
}

type UpdateRoleRequest struct {
	Role string `json:"role" example:"agent" enums:"requester,agent,admin"`
}

// UpdateRole godoc
// @Summary      Ubah peran pengguna (admin)
// @Description  Khusus admin. Admin tidak dapat mengubah perannya sendiri dan admin terakhir tidak dapat diturunkan.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int                true  "User ID"
// @Param        body  body      UpdateRoleRequest  true  "Peran baru"
// @Success      200   {object}  UserEnvelope
// @Failure      400   {object}  response.ErrorEnvelope
// @Failure      401   {object}  response.ErrorEnvelope
// @Failure      403   {object}  response.ErrorEnvelope
// @Failure      404   {object}  response.ErrorEnvelope
// @Failure      409   {object}  response.ErrorEnvelope
// @Router       /users/{id}/role [patch]
func (h *UserHandler) UpdateRole(c *gin.Context) {
	id, appErr := parseIDParam(c, "id")
	if appErr != nil {
		response.ErrorFrom(c, appErr)
		return
	}
	var req UpdateRoleRequest
	if err := response.BindJSON(c, &req); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	user, err := h.users.UpdateRole(c.Request.Context(), middleware.CurrentUser(c), id, req.Role)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.OK(c, toUserResponse(user))
}
