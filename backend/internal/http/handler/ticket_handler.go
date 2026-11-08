package handler

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"

	"helpdesk/backend/internal/domain"
	"helpdesk/backend/internal/http/middleware"
	"helpdesk/backend/internal/http/response"
	"helpdesk/backend/internal/service"
)

type TicketHandler struct {
	tickets *service.TicketService
}

func NewTicketHandler(tickets *service.TicketService) *TicketHandler {
	return &TicketHandler{tickets: tickets}
}

type createTicketRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Priority    string `json:"priority"`
}

func (h *TicketHandler) Create(c *gin.Context) {
	var req createTicketRequest
	if err := response.BindJSON(c, &req); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	ticket, err := h.tickets.Create(c.Request.Context(), middleware.CurrentUser(c), service.CreateTicketInput{
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Priority:    req.Priority,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, toTicketResponse(ticket))
}

func (h *TicketHandler) List(c *gin.Context) {
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

	filter := domain.TicketFilter{
		Page:    page,
		PerPage: perPage,
		Search:  c.Query("search"),
		Sort:    strings.TrimSpace(c.Query("sort")),
		Order:   strings.TrimSpace(c.Query("order")),
	}

	if raw := strings.TrimSpace(c.Query("status")); raw != "" {
		status, ok := domain.ParseStatus(raw)
		if !ok {
			response.ErrorFrom(c, domain.Validation(map[string]string{"status": "Status tidak valid."}))
			return
		}
		filter.Status = &status
	}
	if raw := strings.TrimSpace(c.Query("priority")); raw != "" {
		priority, ok := domain.ParsePriority(raw)
		if !ok {
			response.ErrorFrom(c, domain.Validation(map[string]string{"priority": "Prioritas tidak valid."}))
			return
		}
		filter.Priority = &priority
	}
	filter.AssigneeID, appErr = parseOptionalIDQuery(c, "assignee_id")
	if appErr != nil {
		response.ErrorFrom(c, appErr)
		return
	}
	filter.Unassigned, appErr = parseBoolQuery(c, "unassigned")
	if appErr != nil {
		response.ErrorFrom(c, appErr)
		return
	}

	tickets, total, err := h.tickets.List(c.Request.Context(), middleware.CurrentUser(c), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	items := make([]ticketResponse, 0, len(tickets))
	for _, ticket := range tickets {
		items = append(items, toTicketResponse(ticket))
	}
	response.List(c, items, response.MetaFrom(page, perPage, total))
}

type ticketDetailResponse struct {
	ticketResponse
	Comments []commentResponse `json:"comments"`
}

func (h *TicketHandler) Get(c *gin.Context) {
	id, appErr := parseIDParam(c, "id")
	if appErr != nil {
		response.ErrorFrom(c, appErr)
		return
	}
	detail, err := h.tickets.Get(c.Request.Context(), middleware.CurrentUser(c), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.OK(c, ticketDetailResponse{
		ticketResponse: toTicketResponse(detail.Ticket),
		Comments:       toCommentResponses(detail.Comments),
	})
}

type updateTicketRequest struct {
	Status     *string         `json:"status"`
	Priority   *string         `json:"priority"`
	AssigneeID json.RawMessage `json:"assignee_id"`
}

func (h *TicketHandler) Update(c *gin.Context) {
	id, appErr := parseIDParam(c, "id")
	if appErr != nil {
		response.ErrorFrom(c, appErr)
		return
	}

	var req updateTicketRequest
	if err := response.BindJSON(c, &req); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	patch := domain.TicketUpdate{}
	if req.Status != nil {
		status, ok := domain.ParseStatus(*req.Status)
		if !ok {
			response.ErrorFrom(c, domain.Validation(map[string]string{"status": "Status tidak valid."}))
			return
		}
		patch.Status = &status
	}
	if req.Priority != nil {
		priority, ok := domain.ParsePriority(*req.Priority)
		if !ok {
			response.ErrorFrom(c, domain.Validation(map[string]string{"priority": "Prioritas tidak valid."}))
			return
		}
		patch.Priority = &priority
	}
	if len(req.AssigneeID) > 0 {
		patch.AssigneeSet = true
		raw := strings.TrimSpace(string(req.AssigneeID))
		if raw != "null" {
			var assigneeID int64
			if err := json.Unmarshal(req.AssigneeID, &assigneeID); err != nil || assigneeID <= 0 {
				response.ErrorFrom(c, domain.Validation(map[string]string{"assignee_id": "assignee_id tidak valid."}))
				return
			}
			patch.AssigneeID = &assigneeID
		}
	}

	ticket, err := h.tickets.Update(c.Request.Context(), middleware.CurrentUser(c), id, patch)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.OK(c, toTicketResponse(ticket))
}
