package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	di "github.com/pozedorum/set_pr_reviers_service/internal/DI"
	"github.com/pozedorum/set_pr_reviers_service/pkg/config"
)

func main() {
	cfg := config.Load()
	aplicationContainer, err := di.NewContainer(cfg)
	if err != nil {
		fmt.Println("error while loading container: ", err)
		return
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := aplicationContainer.Start(); err != nil {
			fmt.Println(err)
			return
		}
	}()

	<-quit
	fmt.Println("Shutting down server...")

	if err := aplicationContainer.Shutdown(); err != nil {
		fmt.Println(err)
	}
}
