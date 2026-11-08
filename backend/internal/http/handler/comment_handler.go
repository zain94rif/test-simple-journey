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

type CreateCommentRequest struct {
	Body       string `json:"body" example:"Sedang kami cek, mohon ditunggu."`
	IsInternal bool   `json:"is_internal" example:"false"`
}

// Create godoc
// @Summary      Tambah komentar atau catatan internal
// @Description  Requester hanya boleh membuat komentar publik pada tiketnya sendiri. Catatan internal (is_internal=true) khusus agent dan admin.
// @Tags         comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int                   true  "Ticket ID"
// @Param        body  body      CreateCommentRequest  true  "Isi komentar"
// @Success      201   {object}  CommentEnvelope
// @Failure      400   {object}  response.ErrorEnvelope
// @Failure      401   {object}  response.ErrorEnvelope
// @Failure      403   {object}  response.ErrorEnvelope
// @Failure      404   {object}  response.ErrorEnvelope
// @Router       /tickets/{id}/comments [post]
func (h *CommentHandler) Create(c *gin.Context) {
	ticketID, appErr := parseIDParam(c, "id")
	if appErr != nil {
		response.ErrorFrom(c, appErr)
		return
	}
	var req CreateCommentRequest
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
