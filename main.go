package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/kotalco/core-api/api"
	"github.com/kotalco/core-api/config"
	"github.com/kotalco/core-api/pkg/middleware"
	"github.com/kotalco/core-api/pkg/migration"
	"github.com/kotalco/core-api/pkg/monitor"
	"github.com/kotalco/core-api/pkg/seeder"
	"github.com/kotalco/core-api/pkg/sqlclient"
	"github.com/kotalco/core-api/server"
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
