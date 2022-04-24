package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/cloud-api/api"
	"github.com/kotalco/cloud-api/pkg/config"
	"github.com/kotalco/cloud-api/pkg/middleware"
	"github.com/kotalco/cloud-api/pkg/migration"
	"github.com/kotalco/cloud-api/pkg/seeder"
	"github.com/kotalco/cloud-api/pkg/server"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
)

func main() {
	app := fiber.New(config.FiberConfig())
	middleware.FiberMiddleware(app)
	api.MapUrl(app)

	dbClient := sqlclient.OpenDBConnection()
	migrationService := migration.NewService(dbClient)
	for _, step := range migrationService.Migrate() {
		if err := step.Run(); err != nil {
			go logger.Error(step.Name, err)
		}
	}
	if config.EnvironmentConf["ENVIRONMENT"] == "development" {
		seederService := seeder.NewService(dbClient)
		for _, step := range seederService.Seed() {
			if err := step.Run(); err != nil {
				go logger.Error(step.Name, err)
			}
		}
	}

	server.StartServerWithGracefulShutdown(app)
}
