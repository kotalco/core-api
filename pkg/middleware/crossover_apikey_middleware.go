package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/pkg/config"
	"net/http"
)

func CrossoverAPIKeyProtected(c *fiber.Ctx) error {
	headers := c.GetReqHeaders()
	apiKey := headers["X-Api-Key"]
	if apiKey != config.Environment.CrossOverAPIKey {
		return c.SendStatus(http.StatusUnauthorized)
	}
	return c.Next()
}
