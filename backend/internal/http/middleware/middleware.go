package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"helpdesk/backend/internal/auth"
	"helpdesk/backend/internal/domain"
	"helpdesk/backend/internal/http/response"
)

const (
	userContextKey      = "auth.user"
	requestIDContextKey = "request_id"
)

func CurrentUser(c *gin.Context) domain.User {
	value, ok := c.Get(userContextKey)
	if !ok {
		return domain.User{}
	}
	user, _ := value.(domain.User)
	return user
}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := strings.TrimSpace(c.GetHeader("X-Request-ID"))
		if id == "" {
			id = newRequestID()
		}
		c.Set(requestIDContextKey, id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

func newRequestID() string {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("req-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}

// Auth verifies the Bearer JWT and reloads the user from the database so role
// changes take effect immediately.
func Auth(jwtManager *auth.Manager, users domain.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			response.ErrorFrom(c, domain.Unauthenticated())
			c.Abort()
			return
		}
		raw := strings.TrimSpace(strings.TrimPrefix(header, prefix))
		userID, _, err := jwtManager.Parse(raw)
		if err != nil {
			response.ErrorFrom(c, domain.Unauthenticated())
			c.Abort()
			return
		}

		user, err := users.GetByID(c.Request.Context(), userID)
		if err != nil {
			var appErr *domain.AppError
			if errors.As(err, &appErr) && appErr.Code == "NOT_FOUND" {
				response.ErrorFrom(c, domain.Unauthenticated())
			} else {
				response.ErrorFrom(c, err)
			}
			c.Abort()
			return
		}
		c.Set(userContextKey, user)
		c.Next()
	}
}

func RequireRole(roles ...domain.Role) gin.HandlerFunc {
	allowed := make(map[domain.Role]bool, len(roles))
	for _, role := range roles {
		allowed[role] = true
	}
	return func(c *gin.Context) {
		user := CurrentUser(c)
		if !allowed[user.Role] {
			response.ErrorFrom(c, domain.Forbidden("Anda tidak memiliki akses ke sumber daya ini."))
			c.Abort()
			return
		}
		c.Next()
	}
}

func CORS(origins []string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(origins))
	allowAll := false
	for _, origin := range origins {
		if origin == "*" {
			allowAll = true
		}
		allowed[origin] = true
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && (allowAll || allowed[origin]) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
			c.Header("Access-Control-Expose-Headers", "X-Request-ID")
			c.Header("Access-Control-Max-Age", "600")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}

func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		userID := int64(0)
		if user := CurrentUser(c); user.ID != 0 {
			userID = user.ID
		}
		slog.Info("request",
			"request_id", c.GetString(requestIDContextKey),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"user_id", userID,
			"ip", c.ClientIP(),
		)
	}
}

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		slog.Error("panic tertangkap",
			"panic", recovered,
			"request_id", c.GetString(requestIDContextKey),
			"path", c.Request.URL.Path,
		)
		response.ErrorFrom(c, domain.Internal(fmt.Errorf("panic: %v", recovered)))
		c.Abort()
	})
}
