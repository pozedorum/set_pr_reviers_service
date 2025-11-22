package interfaces

import (
	"context"

	"github.com/gin-gonic/gin"
)

type Server interface {
	Start() error
	Shutdown(ctx context.Context) error
	SetupRoutes(router *gin.Engine, apiRouter *gin.RouterGroup)
}
