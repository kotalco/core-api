package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/kotalco/cloud-api/api"
	"github.com/kotalco/cloud-api/pkg/config"
	"github.com/kotalco/cloud-api/pkg/server"
)

func main() {

	config := config.FiberConfig()
	app := fiber.New(config)
	app.Use(recover.New())
	app.Use(cors.New())
	api.MapUrl(app)

	server.StartServerWithGracefulShutdown(app)

}
