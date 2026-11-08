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
	items := make([]userResponse, 0, len(users))
	for _, user := range users {
		items = append(items, toUserResponse(user))
	}
	response.List(c, items, response.MetaFrom(page, perPage, total))
}

type updateRoleRequest struct {
	Role string `json:"role"`
}

func (h *UserHandler) UpdateRole(c *gin.Context) {
	id, appErr := parseIDParam(c, "id")
	if appErr != nil {
		response.ErrorFrom(c, appErr)
		return
	}
	var req updateRoleRequest
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
