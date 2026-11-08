package handler

import (
	"time"

	"github.com/gin-gonic/gin"

	"helpdesk/backend/internal/domain"
	"helpdesk/backend/internal/http/middleware"
	"helpdesk/backend/internal/http/response"
	"helpdesk/backend/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type RegisterRequest struct {
	Name     string `json:"name" example:"Budi Santoso"`
	Email    string `json:"email" example:"budi@example.com"`
	Password string `json:"password" example:"Rahasia123"`
}

// Register godoc
// @Summary      Registrasi pengguna baru
// @Description  Membuat akun baru. Peran akun baru selalu "requester".
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      RegisterRequest  true  "Data registrasi"
// @Success      201   {object}  UserEnvelope
// @Failure      400   {object}  response.ErrorEnvelope
// @Failure      409   {object}  response.ErrorEnvelope
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := response.BindJSON(c, &req); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	user, err := h.auth.Register(c.Request.Context(), service.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, toUserResponse(user))
}

type LoginRequest struct {
	Email    string `json:"email" example:"admin@helpdesk.local"`
	Password string `json:"password" example:"Admin123!"`
}

type LoginResponse struct {
	Token     string       `json:"token"`
	TokenType string       `json:"token_type" example:"Bearer"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      UserResponse `json:"user"`
}

// Login godoc
// @Summary      Login dan dapatkan JWT
// @Description  Menukar email dan password dengan token JWT HS256.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      LoginRequest  true  "Kredensial login"
// @Success      200   {object}  LoginEnvelope
// @Failure      400   {object}  response.ErrorEnvelope
// @Failure      401   {object}  response.ErrorEnvelope
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := response.BindJSON(c, &req); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	result, err := h.auth.Login(c.Request.Context(), service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.OK(c, LoginResponse{
		Token:     result.Token,
		TokenType: "Bearer",
		ExpiresAt: result.ExpiresAt,
		User:      toUserResponse(result.User),
	})
}

// Me godoc
// @Summary      Profil pengguna saat ini
// @Description  Mengembalikan pengguna yang terautentikasi. Peran dibaca ulang dari database pada setiap request.
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  UserEnvelope
// @Failure      401  {object}  response.ErrorEnvelope
// @Router       /me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user.ID == 0 {
		response.ErrorFrom(c, domain.Unauthenticated())
		return
	}
	response.OK(c, toUserResponse(user))
}
