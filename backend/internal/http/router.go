package http

import (
	"github.com/gin-gonic/gin"

	"helpdesk/backend/internal/auth"
	"helpdesk/backend/internal/config"
	"helpdesk/backend/internal/domain"
	"helpdesk/backend/internal/http/handler"
	"helpdesk/backend/internal/http/middleware"
	"helpdesk/backend/internal/http/response"
)

type Dependencies struct {
	Config   config.Config
	JWT      *auth.Manager
	Users    domain.UserRepository
	Auth     *handler.AuthHandler
	Tickets  *handler.TicketHandler
	Comments *handler.CommentHandler
	UsersH   *handler.UserHandler
	Health   *handler.HealthHandler
}

func NewRouter(deps Dependencies) *gin.Engine {
	if deps.Config.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.HandleMethodNotAllowed = true
	router.Use(
		middleware.RequestID(),
		middleware.CORS(deps.Config.CORSOrigins),
		middleware.BodyLimit(deps.Config.MaxBodyBytes),
		middleware.AccessLog(),
		middleware.Recovery(),
	)

	router.NoRoute(func(c *gin.Context) {
		response.ErrorFrom(c, domain.NotFound("Endpoint"))
	})
	router.NoMethod(func(c *gin.Context) {
		c.JSON(405, gin.H{"error": gin.H{
			"code":    "METHOD_NOT_ALLOWED",
			"message": "Method tidak diizinkan untuk endpoint ini.",
		}})
	})

	api := router.Group("/api")
	api.GET("/health", deps.Health.Health)
	api.POST("/auth/register", deps.Auth.Register)
	api.POST("/auth/login", deps.Auth.Login)

	protected := api.Group("")
	protected.Use(middleware.Auth(deps.JWT, deps.Users))
	protected.GET("/me", deps.Auth.Me)
	protected.GET("/tickets", deps.Tickets.List)
	protected.POST("/tickets", deps.Tickets.Create)
	protected.GET("/tickets/:id", deps.Tickets.Get)
	protected.PATCH("/tickets/:id", deps.Tickets.Update)
	protected.POST("/tickets/:id/comments", deps.Comments.Create)

	admin := protected.Group("")
	admin.Use(middleware.RequireRole(domain.RoleAdmin))
	admin.GET("/users", deps.UsersH.List)
	admin.PATCH("/users/:id/role", deps.UsersH.UpdateRole)

	return router
}
