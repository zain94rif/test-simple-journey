package handler

import (
	"github.com/gin-gonic/gin"

	"helpdesk/backend/internal/http/middleware"
	"helpdesk/backend/internal/http/response"
	"helpdesk/backend/internal/service"
)

type CommentHandler struct {
	comments *service.CommentService
}

func NewCommentHandler(comments *service.CommentService) *CommentHandler {
	return &CommentHandler{comments: comments}
}

type createCommentRequest struct {
	Body       string `json:"body"`
	IsInternal bool   `json:"is_internal"`
}

func (h *CommentHandler) Create(c *gin.Context) {
	ticketID, appErr := parseIDParam(c, "id")
	if appErr != nil {
		response.ErrorFrom(c, appErr)
		return
	}
	var req createCommentRequest
	if err := response.BindJSON(c, &req); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	comment, err := h.comments.Create(c.Request.Context(), middleware.CurrentUser(c), ticketID, service.CreateCommentInput{
		Body:       req.Body,
		IsInternal: req.IsInternal,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, toCommentResponse(comment))
}
