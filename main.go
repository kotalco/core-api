package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/cloud-api/api"
	"github.com/kotalco/cloud-api/internal/user"
	"github.com/kotalco/cloud-api/internal/verification"
	"github.com/kotalco/cloud-api/pkg/config"
	"github.com/kotalco/cloud-api/pkg/middleware"
	"github.com/kotalco/cloud-api/pkg/server"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
)

func main() {

	config := config.FiberConfig()
	app := fiber.New(config)
	middleware.FiberMiddleware(app)
	api.MapUrl(app)

	dbClient, err := sqlclient.OpenDBConnection()
	if err != nil {
		go logger.Error("DATABASE_FAILED_CONNECTION", err)
		return
	}

	err = dbClient.AutoMigrate(new(user.User))
	err = dbClient.AutoMigrate(new(verification.Verification))
	if err != nil {
		go logger.Error("MIGRATION_FAILED", err)
		return
	}

	server.StartServerWithGracefulShutdown(app)
}
