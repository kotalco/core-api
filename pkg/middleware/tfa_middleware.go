package middleware

import (
	"github.com/gofiber/fiber/v2"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/cloud-api/pkg/tokens"
)

func TFAProtected(c *fiber.Ctx) error {
	BearerToken := c.Get("Authorization")
	accessDetails, err := tokens.ExtractTokenMetadata(BearerToken)
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
