package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pozedorum/set_pr_reviers_service/internal/interfaces"
)

type PRServer struct {
	server *http.Server
	router *gin.Engine
	serv   interfaces.Service
	logger interfaces.Logger
}

func NewPRServer(port string, service interfaces.Service, logger interfaces.Logger) *PRServer {
	router := gin.Default()

	s := &PRServer{
		serv:   service,
		logger: logger,
		server: &http.Server{
			Addr:    ":" + port,
			Handler: router,
		},
		router: router,
	}

	s.setupRoutes()
	return s
}

func (s *PRServer) setupRoutes() {
	s.router.GET("/health", s.handleHealthCheck)

	// Teams group
	teams := s.router.Group("/team")
	{
		teams.POST("/add", s.handleCreateTeam) // POST /team/add
		teams.GET("/get", s.handleGetTeam)     // GET /team/get
	}

	// Users group
	users := s.router.Group("/users")
	{
		users.POST("/setIsActive", s.handleSetUserActive) // POST /users/setIsActive
		users.GET("/getReview", s.handleGetUserReviews)   // GET /users/getReview
	}

	// Pull Requests group
	prs := s.router.Group("/pullRequest")
	{
		prs.POST("/create", s.handleCreatePR)           // POST /pullRequest/create
		prs.POST("/merge", s.handleMergePR)             // POST /pullRequest/merge
		prs.POST("/reassign", s.handleReassignReviewer) // POST /pullRequest/reassign
	}
}

func (s *PRServer) Start() error {
	s.logger.Info("SERVER_START", "Gin server starting", "addr", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *PRServer) Shutdown(ctx context.Context) error {
	s.logger.Info("SERVER_SHUTDOWN", "Initiating server shutdown")
	return s.server.Shutdown(ctx)
}

func (s *PRServer) handleHealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}
