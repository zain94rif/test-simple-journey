package handler

import "helpdesk/backend/internal/http/response"

// The types below exist so Swagger can describe the actual response envelopes
// without duplicating the JSON shape used by the handlers.

// UserEnvelope wraps a single user.
type UserEnvelope struct {
	Data UserResponse `json:"data"`
}

// LoginEnvelope wraps the login payload.
type LoginEnvelope struct {
	Data LoginResponse `json:"data"`
}

// TicketEnvelope wraps a single ticket.
type TicketEnvelope struct {
	Data TicketResponse `json:"data"`
}

// TicketDetailEnvelope wraps a ticket together with its comments.
type TicketDetailEnvelope struct {
	Data TicketDetailResponse `json:"data"`
}

// CommentEnvelope wraps a single comment.
type CommentEnvelope struct {
	Data CommentResponse `json:"data"`
}

// TicketListEnvelope is the paginated ticket list response.
type TicketListEnvelope struct {
	Data []TicketResponse `json:"data"`
	Meta response.Meta    `json:"meta"`
}

// UserListEnvelope is the paginated user list response.
type UserListEnvelope struct {
	Data []UserResponse `json:"data"`
	Meta response.Meta  `json:"meta"`
}

// HealthResponse is the health check payload.
type HealthResponse struct {
	Status string `json:"status" example:"ok"`
	DB     string `json:"db" example:"up"`
}
