package monitor

import (
	"github.com/kotalco/cloud-api/config"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
)

func NewMonitor(a *fiber.App) {
	if config.Environment.Environment == "development" {
		a.Get("/metrics", monitor.New(conf()))
	}
}

func conf() monitor.Config {
	return monitor.Config{
		Title:      "Kotal Monitor",
		Refresh:    3 * time.Second,
		APIOnly:    false,
		Next:       nil,
		CustomHead: "",
	}
}
