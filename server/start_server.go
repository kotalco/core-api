package server

import (
	"fmt"
	"github.com/kotalco/cloud-api/config"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/pkg/logger"
)

// StartServerWithGracefulShutdown function for starting server with a graceful shutdown.
// Create channel for idle connections.
// check if  Received an interrupt signal, shutdown.
// Error if closing listeners, or context timeout
// Run server.
// Error if  Run server with reason
func StartServerWithGracefulShutdown(a *fiber.App) {
	idleConnsClosed := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT) // Catch OS signals.
		<-sigint
		if err := a.Shutdown(); err != nil {
			go logger.Info("StartServerWithGracefulShutdown", fmt.Sprintf("Oops... Server is not shutting down! Reason:  %v", err))
		}
		close(idleConnsClosed)
	}()

	port := ":" + config.Environment.ServerPort

	if err := a.Listen(port); err != nil {
		go logger.Info("StartServerWithGracefulShutdown", fmt.Sprintf("Oops... Server is not running! Reason: %v", err))
	}
	<-idleConnsClosed
}
