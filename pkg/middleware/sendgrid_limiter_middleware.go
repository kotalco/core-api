package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	restError "github.com/kotalco/core-api/pkg/errors"
	"time"
)

func SendGridLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        1,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			err := restError.NewTooManyRequestsError("You have reached the maximum number of verification email requests allowed per minute")
			return c.Status(err.StatusCode()).JSON(err)
		},
	})
}
