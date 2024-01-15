package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/kotalco/core-api/config"
)

// FiberMiddleware provide Fiber's built-in middlewares.
func FiberMiddleware(a *fiber.App) {
	a.Use(
		recover.New(),
		cors.New(),
		config.FiberLimiter(),
	)
}
