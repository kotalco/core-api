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
	app := fiber.New(server.FiberConfig())
	middleware.FiberMiddleware(app)
	monitor.NewMonitor(app)
	app.Use(pprof.New())
	api.MapUrl(app)

	k8s.Config()

	//open db connection
	dbClient := sqlclient.OpenDBConnection()
	migrationService := migration.NewService(dbClient)
	for _, step := range migrationService.Migrations() {
		if err := step.Run(); err != nil {
			go logger.Error(step.Name, err)
		}
	}

	//seed production tables
	seederService := seeder.NewService(dbClient)
	err := seederService.Seeds()[seeder.SeedSettingTable].Run()
	if err != nil {
		go logger.Error(seeder.SeedSettingTable, err)
	}
	//seed developments tables
	if config.Environment.Environment == "development" {
		for v, step := range seederService.Seeds() {
			if err := step.Run(); err != nil {
				go logger.Error(v, err)
			}
		}
	}

	server.StartServerWithGracefulShutdown(app)
}
