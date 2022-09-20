package monitor

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/kotalco/cloud-api/pkg/config"
	"time"
)

func NewMonitor(a *fiber.App) {
	if config.EnvironmentConf["ENVIRONMENT"] == "development" {
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
