package httpapi

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"taskflow/backend/internal/store"
)

type Handler struct {
	store      *store.Store
	logger     *slog.Logger
	jwtSecret  string
	jwtExpiry  time.Duration
	bcryptCost int
}

func NewHandler(store *store.Store, logger *slog.Logger, jwtSecret string, jwtExpiry time.Duration, bcryptCost int) *Handler {
	return &Handler{
		store:      store,
		logger:     logger,
		jwtSecret:  jwtSecret,
		jwtExpiry:  jwtExpiry,
		bcryptCost: bcryptCost,
	}
}

func NewRouter(handler *Handler, logger *slog.Logger, jwtSecret string, allowedOrigin string) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(JSONContentTypeMiddleware())
	router.Use(LoggingMiddleware(logger))
	router.Use(CORSMiddleware(allowedOrigin))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.POST("/auth/register", handler.Register)
	router.POST("/auth/login", handler.Login)

	authed := router.Group("/")
	authed.Use(AuthMiddleware(jwtSecret))
	{
		authed.GET("/users", handler.ListUsers)

		authed.GET("/projects", handler.ListProjects)
		authed.POST("/projects", handler.CreateProject)
		authed.GET("/projects/:id", handler.GetProject)
		authed.PATCH("/projects/:id", handler.UpdateProject)
		authed.DELETE("/projects/:id", handler.DeleteProject)
		authed.GET("/projects/:id/tasks", handler.ListTasks)
		authed.POST("/projects/:id/tasks", handler.CreateTask)
		authed.GET("/projects/:id/stats", handler.ProjectStats)

		authed.PATCH("/tasks/:id", handler.UpdateTask)
		authed.DELETE("/tasks/:id", handler.DeleteTask)
	}

	return router
}
