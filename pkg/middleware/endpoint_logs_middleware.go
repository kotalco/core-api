package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/pkg/config"
	"net/http"
)

func EndpointLogsAPIKeyProtected(c *fiber.Ctx) error {
	headers := c.GetReqHeaders()
	apiKey := headers["X-Api-Key"]
	if apiKey != config.Environment.EndpointLogsAPIKey {
		return c.SendStatus(http.StatusUnauthorized)
	}
	return c.Next()
}
