package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/kotalco/cloud-api/api"
	"github.com/kotalco/cloud-api/pkg/config"
	"github.com/kotalco/cloud-api/pkg/k8s"
	"github.com/kotalco/cloud-api/pkg/middleware"
	"github.com/kotalco/cloud-api/pkg/migration"
	"github.com/kotalco/cloud-api/pkg/monitor"
	"github.com/kotalco/cloud-api/pkg/seeder"
	"github.com/kotalco/cloud-api/pkg/server"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	"github.com/kotalco/community-api/pkg/logger"
)

func main() {
	app := fiber.New(config.FiberConfig())
	middleware.FiberMiddleware(app)
	monitor.NewMonitor(app)
	app.Use(pprof.New())
	api.MapUrl(app)

	//Adds additional types to the community k8s client
	k8s.AddToScheme()
	//open db connection
	dbClient := sqlclient.OpenDBConnection()
	migrationService := migration.NewService(dbClient)
	for _, step := range migrationService.Migrations() {
		if err := step.Run(); err != nil {
			go logger.Error(step.Name, err)
		}
	}
	if config.EnvironmentConf["ENVIRONMENT"] == "development" {
		seederService := seeder.NewService(dbClient)
		for _, step := range seederService.Seeds() {
			if err := step.Run(); err != nil {
				go logger.Error(step.Name, err)
			}
		}
	}

	server.StartServerWithGracefulShutdown(app)
}
