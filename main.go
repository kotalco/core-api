package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/cloud-api/api"
	"github.com/kotalco/cloud-api/pkg/config"
	"github.com/kotalco/cloud-api/pkg/middleware"
	"github.com/kotalco/cloud-api/pkg/migration"
	"github.com/kotalco/cloud-api/pkg/server"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
)

func main() {
	app := fiber.New(config.FiberConfig())
	middleware.FiberMiddleware(app)
	api.MapUrl(app)

	dbClient := sqlclient.OpenDBConnection()
	if config.EnvironmentConf["ENVIRONMENT"] == "development" {
		migrationService := migration.NewService(dbClient)
		for _, definition := range migrationService.Migrate() {
			if err := definition.Run(); err != nil {
				go logger.Error(definition.Name, err)
			}
		}
	}

	server.StartServerWithGracefulShutdown(app)
}
