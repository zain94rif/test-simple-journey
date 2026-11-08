package domain

import (
	"fmt"
	"net/http"
)

// AppError is an error that can be safely returned to API clients.
type AppError struct {
	Status  int
	Code    string
	Message string
	Details map[string]string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

func Validation(details map[string]string) *AppError {
	return &AppError{
		Status:  http.StatusBadRequest,
		Code:    "VALIDATION_ERROR",
		Message: "Data yang dikirim tidak valid.",
		Details: details,
	}
}

func ValidationMessage(message string) *AppError {
	return &AppError{
		Status:  http.StatusBadRequest,
		Code:    "VALIDATION_ERROR",
		Message: message,
	}
}

func Unauthenticated() *AppError {
	return &AppError{
		Status:  http.StatusUnauthorized,
		Code:    "UNAUTHENTICATED",
		Message: "Sesi tidak valid atau telah berakhir.",
	}
}

func InvalidCredentials() *AppError {
	return &AppError{
		Status:  http.StatusUnauthorized,
		Code:    "INVALID_CREDENTIALS",
		Message: "Email atau password salah.",
	}
}

func Forbidden(message string) *AppError {
	if message == "" {
		message = "Anda tidak memiliki akses untuk tindakan ini."
	}
	return &AppError{
		Status:  http.StatusForbidden,
		Code:    "FORBIDDEN",
		Message: message,
	}
}

func NotFound(resource string) *AppError {
	return &AppError{
		Status:  http.StatusNotFound,
		Code:    "NOT_FOUND",
		Message: fmt.Sprintf("%s tidak ditemukan.", resource),
	}
}

func DuplicateEmail() *AppError {
	return &AppError{
		Status:  http.StatusConflict,
		Code:    "DUPLICATE_EMAIL",
		Message: "Email sudah terdaftar.",
	}
}

func InvalidTransition(from, to Status, role Role) *AppError {
	return &AppError{
		Status:  http.StatusConflict,
		Code:    "INVALID_TRANSITION",
		Message: fmt.Sprintf("Transisi dari '%s' ke '%s' tidak diizinkan untuk peran '%s'.", from, to, role),
		Details: map[string]string{
			"from": string(from),
			"to":   string(to),
			"role": string(role),
		},
	}
}

func InternalNoteForbidden() *AppError {
	return &AppError{
		Status:  http.StatusForbidden,
		Code:    "INTERNAL_NOTE_FORBIDDEN",
		Message: "Hanya agent dan admin yang dapat membuat catatan internal.",
	}
}

func SelfRoleChange() *AppError {
	return &AppError{
		Status:  http.StatusConflict,
		Code:    "SELF_ROLE_CHANGE",
		Message: "Anda tidak dapat mengubah peran Anda sendiri.",
	}
}

func LastAdmin() *AppError {
	return &AppError{
		Status:  http.StatusConflict,
		Code:    "LAST_ADMIN",
		Message: "Peran admin terakhir tidak dapat diturunkan.",
	}
}

func PayloadTooLarge() *AppError {
	return &AppError{
		Status:  http.StatusRequestEntityTooLarge,
		Code:    "PAYLOAD_TOO_LARGE",
		Message: "Ukuran request melebihi batas yang diizinkan.",
	}
}

func Internal(err error) *AppError {
	return &AppError{
		Status:  http.StatusInternalServerError,
		Code:    "INTERNAL_ERROR",
		Message: "Terjadi kesalahan pada server.",
		Err:     err,
	}
}
