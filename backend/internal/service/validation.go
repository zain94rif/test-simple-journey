package service

import (
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"
)

func runeLen(s string) int { return utf8.RuneCountInString(s) }

func normalizeEmail(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func validateEmail(raw string) (string, string) {
	email := normalizeEmail(raw)
	if email == "" {
		return email, "Email wajib diisi."
	}
	if len(email) > 255 {
		return email, "Email maksimal 255 karakter."
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return email, "Format email tidak valid."
	}
	return email, ""
}

func validateName(raw string) (string, string) {
	name := strings.TrimSpace(raw)
	if runeLen(name) < 2 {
		return name, "Nama minimal 2 karakter."
	}
	if runeLen(name) > 100 {
		return name, "Nama maksimal 100 karakter."
	}
	return name, ""
}

func validatePassword(raw string) string {
	if len(raw) < 8 {
		return "Password minimal 8 karakter."
	}
	if len(raw) > 72 {
		return "Password maksimal 72 karakter."
	}
	var hasLetter, hasDigit bool
	for _, r := range raw {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return "Password harus mengandung huruf dan angka."
	}
	return ""
}

func validateTitle(raw string) (string, string) {
	title := strings.TrimSpace(raw)
	if runeLen(title) < 3 {
		return title, "Judul minimal 3 karakter."
	}
	if runeLen(title) > 150 {
		return title, "Judul maksimal 150 karakter."
	}
	return title, ""
}

func validateDescription(raw string) (string, string) {
	description := strings.TrimSpace(raw)
	if runeLen(description) < 10 {
		return description, "Deskripsi minimal 10 karakter."
	}
	if runeLen(description) > 5000 {
		return description, "Deskripsi maksimal 5000 karakter."
	}
	return description, ""
}

func validateCommentBody(raw string) (string, string) {
	body := strings.TrimSpace(raw)
	if body == "" {
		return body, "Komentar tidak boleh kosong."
	}
	if runeLen(body) > 5000 {
		return body, "Komentar maksimal 5000 karakter."
	}
	return body, ""
}
