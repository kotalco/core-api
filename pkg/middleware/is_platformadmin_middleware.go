package middleware

import (
	"github.com/gofiber/fiber/v2"
	restErrors "github.com/kotalco/cloud-api/pkg/errors"
	"github.com/kotalco/cloud-api/pkg/token"
)

func IsPlatformAdmin(c *fiber.Ctx) error {
	if !c.Locals("user").(token.UserDetails).PlatformAdmin {
		forbidden := restErrors.NewForbiddenError("unAuthorized action")
		return c.Status(forbidden.StatusCode()).JSON(forbidden)
	}
	return c.Next()
}
