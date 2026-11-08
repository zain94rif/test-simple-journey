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

type CreateTicketRequest struct {
	Title       string `json:"title" example:"Laptop tidak menyala"`
	Description string `json:"description" example:"Laptop mati total setelah update sistem."`
	Category    string `json:"category" example:"hardware" enums:"hardware,software,network,access,other"`
	Priority    string `json:"priority" example:"high" enums:"low,medium,high,urgent"`
}

// Create godoc
// @Summary      Buat tiket baru
// @Description  Requester membuat tiket. Status selalu "open" dan requester diambil dari token.
// @Tags         tickets
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      CreateTicketRequest  true  "Data tiket"
// @Success      201   {object}  TicketEnvelope
// @Failure      400   {object}  response.ErrorEnvelope
// @Failure      401   {object}  response.ErrorEnvelope
// @Router       /tickets [post]
func (h *TicketHandler) Create(c *gin.Context) {
	var req CreateTicketRequest
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

// List godoc
// @Summary      Daftar tiket
// @Description  Requester hanya melihat tiket miliknya; agent dan admin melihat semua tiket. Mendukung filter, pencarian judul, pagination, dan sorting.
// @Tags         tickets
// @Produce      json
// @Security     BearerAuth
// @Param        page         query     int     false  "Halaman (default 1)"
// @Param        per_page     query     int     false  "Jumlah per halaman (1-100, default 20)"
// @Param        status       query     string  false  "Filter status"  Enums(open, in_progress, resolved, closed)
// @Param        priority     query     string  false  "Filter prioritas"  Enums(low, medium, high, urgent)
// @Param        assignee_id  query     int     false  "Filter assignee"
// @Param        unassigned   query     bool    false  "Hanya tiket tanpa assignee"
// @Param        search       query     string  false  "Cari pada judul tiket"
// @Param        sort         query     string  false  "Kolom sort"  Enums(created_at, updated_at, priority, status)
// @Param        order        query     string  false  "Arah sort"  Enums(asc, desc)
// @Success      200  {object}  TicketListEnvelope
// @Failure      400  {object}  response.ErrorEnvelope
// @Failure      401  {object}  response.ErrorEnvelope
// @Router       /tickets [get]
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
	items := make([]TicketResponse, 0, len(tickets))
	for _, ticket := range tickets {
		items = append(items, toTicketResponse(ticket))
	}
	response.List(c, items, response.MetaFrom(page, perPage, total))
}

type TicketDetailResponse struct {
	TicketResponse
	Comments []CommentResponse `json:"comments"`
}

// Get godoc
// @Summary      Detail tiket dan komentar
// @Description  Requester hanya dapat mengakses tiketnya sendiri. Catatan internal (is_internal=true) hanya dikembalikan untuk agent dan admin.
// @Tags         tickets
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Ticket ID"
// @Success      200  {object}  TicketDetailEnvelope
// @Failure      401  {object}  response.ErrorEnvelope
// @Failure      403  {object}  response.ErrorEnvelope
// @Failure      404  {object}  response.ErrorEnvelope
// @Router       /tickets/{id} [get]
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
	response.OK(c, TicketDetailResponse{
		TicketResponse: toTicketResponse(detail.Ticket),
		Comments:       toCommentResponses(detail.Comments),
	})
}

type UpdateTicketRequest struct {
	Status     *string         `json:"status" example:"in_progress"`
	Priority   *string         `json:"priority" example:"urgent"`
	AssigneeID json.RawMessage `json:"assignee_id" swaggertype:"integer" example:"2"`
}

// Update godoc
// @Summary      Ubah status, prioritas, atau assignee tiket
// @Description  Hanya transisi status yang diizinkan PRD yang diterima; transisi lain menghasilkan 409 INVALID_TRANSITION. Agent hanya dapat assign dirinya sendiri, admin dapat assign agent/admin. Kirim assignee_id: null untuk melepas assignee.
// @Tags         tickets
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int                  true  "Ticket ID"
// @Param        body  body      UpdateTicketRequest  true  "Field yang diubah"
// @Success      200   {object}  TicketEnvelope
// @Failure      400   {object}  response.ErrorEnvelope
// @Failure      401   {object}  response.ErrorEnvelope
// @Failure      403   {object}  response.ErrorEnvelope
// @Failure      404   {object}  response.ErrorEnvelope
// @Failure      409   {object}  response.ErrorEnvelope
// @Router       /tickets/{id} [patch]
func (h *TicketHandler) Update(c *gin.Context) {
	id, appErr := parseIDParam(c, "id")
	if appErr != nil {
		response.ErrorFrom(c, appErr)
		return
	}

	var req UpdateTicketRequest
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
