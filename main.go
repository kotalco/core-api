package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/kotalco/cloud-api/api"
	"github.com/kotalco/cloud-api/pkg/k8s"
	"github.com/kotalco/cloud-api/pkg/memdb"
	"github.com/kotalco/cloud-api/pkg/middleware"
	"github.com/kotalco/cloud-api/pkg/migration"
	"github.com/kotalco/cloud-api/pkg/monitor"
	"github.com/kotalco/cloud-api/pkg/seeder"
	"github.com/kotalco/cloud-api/pkg/server"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
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
	migrationService.Run()

	seederService := seeder.NewService(dbClient)
	seederService.Run()

	memdb.SeedMemDB()

	server.StartServerWithGracefulShutdown(app)
}
