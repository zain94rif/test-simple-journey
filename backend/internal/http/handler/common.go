package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"helpdesk/backend/internal/domain"
)

type UserResponse struct {
	ID        int64       `json:"id"`
	Name      string      `json:"name"`
	Email     string      `json:"email"`
	Role      domain.Role `json:"role"`
	CreatedAt time.Time   `json:"created_at"`
}

func toUserResponse(user domain.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.UTC(),
	}
}

type TicketUserResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func toTicketUserResponse(user *domain.TicketUser) *TicketUserResponse {
	if user == nil {
		return nil
	}
	return &TicketUserResponse{ID: user.ID, Name: user.Name, Email: user.Email}
}

type TicketResponse struct {
	ID          int64               `json:"id"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Category    domain.Category     `json:"category"`
	Priority    domain.Priority     `json:"priority"`
	Status      domain.Status       `json:"status"`
	Requester   *TicketUserResponse `json:"requester"`
	Assignee    *TicketUserResponse `json:"assignee"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

func toTicketResponse(ticket domain.Ticket) TicketResponse {
	return TicketResponse{
		ID:          ticket.ID,
		Title:       ticket.Title,
		Description: ticket.Description,
		Category:    ticket.Category,
		Priority:    ticket.Priority,
		Status:      ticket.Status,
		Requester:   toTicketUserResponse(ticket.Requester),
		Assignee:    toTicketUserResponse(ticket.Assignee),
		CreatedAt:   ticket.CreatedAt.UTC(),
		UpdatedAt:   ticket.UpdatedAt.UTC(),
	}
}

type CommentResponse struct {
	ID         int64         `json:"id"`
	TicketID   int64         `json:"ticket_id"`
	Author     CommentAuthor `json:"author"`
	Body       string        `json:"body"`
	IsInternal bool          `json:"is_internal"`
	CreatedAt  time.Time     `json:"created_at"`
}

type CommentAuthor struct {
	ID   int64       `json:"id"`
	Name string      `json:"name"`
	Role domain.Role `json:"role"`
}

func toCommentResponse(comment domain.Comment) CommentResponse {
	var author CommentAuthor
	if comment.Author != nil {
		author = CommentAuthor{ID: comment.Author.ID, Name: comment.Author.Name, Role: comment.Author.Role}
	}
	return CommentResponse{
		ID:         comment.ID,
		TicketID:   comment.TicketID,
		Author:     author,
		Body:       comment.Body,
		IsInternal: comment.IsInternal,
		CreatedAt:  comment.CreatedAt.UTC(),
	}
}

func toCommentResponses(comments []domain.Comment) []CommentResponse {
	out := make([]CommentResponse, 0, len(comments))
	for _, comment := range comments {
		out = append(out, toCommentResponse(comment))
	}
	return out
}

func parseIDParam(c *gin.Context, name string) (int64, *domain.AppError) {
	raw := strings.TrimSpace(c.Param(name))
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, domain.Validation(map[string]string{name: "ID tidak valid."})
	}
	return id, nil
}

func parsePositiveIntQuery(c *gin.Context, name string, fallback int) (int, *domain.AppError) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, domain.Validation(map[string]string{name: "Parameter " + name + " harus berupa angka."})
	}
	return value, nil
}

func parseBoolQuery(c *gin.Context, name string) (bool, *domain.AppError) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return false, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, domain.Validation(map[string]string{name: "Parameter " + name + " harus bernilai true atau false."})
	}
	return value, nil
}

func parseOptionalIDQuery(c *gin.Context, name string) (*int64, *domain.AppError) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return nil, domain.Validation(map[string]string{name: "Parameter " + name + " tidak valid."})
	}
	return &value, nil
}
