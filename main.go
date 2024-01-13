package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/kotalco/cloud-api/api"
	"github.com/kotalco/cloud-api/config"
	"github.com/kotalco/cloud-api/pkg/middleware"
	"github.com/kotalco/cloud-api/pkg/migration"
	"github.com/kotalco/cloud-api/pkg/monitor"
	"github.com/kotalco/cloud-api/pkg/seeder"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	"github.com/kotalco/cloud-api/server"
)

func main() {
	app := fiber.New(config.FiberConfig())
	middleware.FiberMiddleware(app)
	monitor.NewMonitor(app)
	app.Use(pprof.New())
	api.MapUrl(app)

	dbClient := sqlclient.OpenDBConnection()

	migrationService := migration.NewService(dbClient)
	migrationService.Run()

	seederService := seeder.NewService(dbClient)
	seederService.Run()

	server.StartServerWithGracefulShutdown(app)
}
