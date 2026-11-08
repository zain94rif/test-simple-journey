package response

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"helpdesk/backend/internal/domain"
)

type ErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

type ErrorEnvelope struct {
	Error ErrorDetail `json:"error"`
}

type Meta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, gin.H{"data": data})
}

func List(c *gin.Context, data any, meta Meta) {
	c.JSON(http.StatusOK, gin.H{"data": data, "meta": meta})
}

// ErrorFrom maps a domain error to the shared error envelope.
func ErrorFrom(c *gin.Context, err error) {
	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		if appErr.Status >= http.StatusInternalServerError {
			slog.Error("request gagal",
				"error", appErr.Error(),
				"code", appErr.Code,
				"request_id", c.GetString("request_id"),
				"path", c.FullPath(),
			)
		}
		c.JSON(appErr.Status, ErrorEnvelope{Error: ErrorDetail{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		}})
		return
	}

	slog.Error("request gagal tanpa AppError",
		"error", err.Error(),
		"request_id", c.GetString("request_id"),
		"path", c.FullPath(),
	)
	c.JSON(http.StatusInternalServerError, ErrorEnvelope{Error: ErrorDetail{
		Code:    "INTERNAL_ERROR",
		Message: "Terjadi kesalahan pada server.",
	}})
}

// BindJSON decodes a JSON body and maps syntax/size errors to the shared
// error contract.
func BindJSON(c *gin.Context, dst any) *domain.AppError {
	if err := c.ShouldBindJSON(dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return domain.PayloadTooLarge()
		}
		var syntaxErr *json.SyntaxError
		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &syntaxErr) || errors.As(err, &typeErr) {
			return domain.Validation(map[string]string{"body": "Body JSON tidak valid."})
		}
		return domain.Validation(map[string]string{"body": "Body request tidak valid atau kosong."})
	}
	return nil
}

func MetaFrom(page, perPage, total int) Meta {
	totalPages := 0
	if total > 0 && perPage > 0 {
		totalPages = (total + perPage - 1) / perPage
	}
	return Meta{Page: page, PerPage: perPage, Total: total, TotalPages: totalPages}
}
