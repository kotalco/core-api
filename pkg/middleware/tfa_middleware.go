package middleware

import (
	"github.com/gofiber/fiber/v2"
	restErrors "github.com/kotalco/api/pkg/errors"
)

func TFAProtected(c *fiber.Ctx) error {
	//get Authorization token from headers or from qs if it doesn't exit in case of ws connections
	BearerToken := c.Get("Authorization", c.Query("Authorization"))
	accessDetails, err := tokenService.ExtractTokenMetadata(BearerToken)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	if !accessDetails.Authorized {
		unAuthErr := restErrors.NewUnAuthorizedError("2factor auth required")
		return c.Status(unAuthErr.Status).JSON(unAuthErr)
	}
	c.Next()
	return nil
}
