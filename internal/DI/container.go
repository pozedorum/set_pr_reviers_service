package di

import (
	"context"
	"fmt"
	"time"

	"github.com/pozedorum/set_pr_reviers_service/internal/interfaces"
	"github.com/pozedorum/set_pr_reviers_service/pkg/config"
	"github.com/pozedorum/set_pr_reviers_service/pkg/logger"
)

type Container struct {
	//	repo     interfaces.Repository
	service interfaces.Service
	server  interfaces.Server
	logger  interfaces.Logger
}

func NewContainer(cfg *config.Config) (*Container, error) {
	// Инициализируем логгер
	// logger, err := logger.NewLogger("event-service", "")
	logger, err := logger.NewLogger("event-service", "./logs/app.log")
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	logger.Info("CONTAINER_INIT", "Starting application container initialization")

	// Репозиторий
	// repo, err := repository.NewEventRepository(cfg.Database.GetDSN(), logger)
	// if err != nil {
	// 	logger.Error("CONTAINER_INIT", "Failed to create repository", "error", err)
	// 	return nil, err
	// }
	// logger.Info("CONTAINER_INIT", "Repository initialized successfully")

	// Business service
	// service := service.NewEventService(repo, logger)
	logger.Info("CONTAINER_INIT", "Service initialized successfully")

	// HTTP server
	// server := server.NewEventServer(cfg.Server.Port, service, logger)
	// logger.Info("CONTAINER_INIT", "Server initialized successfully")

	return &Container{
		// repo:     repo,
		// service:  service,
		// server:   server,
		logger: logger,
	}, nil
}

func (c *Container) Start() error {
	return c.server.Start()
}

func (c *Container) Shutdown() error {
	var errors []error
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := c.server.Shutdown(ctx); err != nil {
		errors = append(errors, fmt.Errorf("server shutdown: %w", err))
	}
	// Shutdown repository
	// if err := c.repo.Close(); err != nil {
	// 	errors = append(errors, fmt.Errorf("repository close: %w", err))
	// }

	// Shutdown logger
	c.logger.Shutdown()

	if len(errors) > 0 {
		return fmt.Errorf("shutdown completed with errors: %v", errors)
	}

	c.logger.Info("CONTAINER_SHUTDOWN", "Container shutdown completed successfully")
	return nil
}
