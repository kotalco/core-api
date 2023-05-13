package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/kotalco/cloud-api/pkg/config"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	"github.com/kotalco/community-api/pkg/shared"

	"strconv"
	"time"
)

func FiberConfig() fiber.Config {
	readTimeoutSecondsCount, _ := strconv.Atoi(config.Environment.ServerReadTimeout)
	return fiber.Config{
		ReadTimeout:  time.Second * time.Duration(readTimeoutSecondsCount),
		ErrorHandler: defaultErrorHandler,
	}
}

// defaultErrorHandler used to catch all unhandled  run time resterror mainly panics
// logs resterror using logger pkg
// return custom error struct using restError pkg
var defaultErrorHandler = func(c *fiber.Ctx, err error) error {
	go logger.Warn("PANICKING_UNHANDLED_ERROR", err)
	internalErr := restErrors.NewInternalServerError("some thing went wrong")

	return c.Status(internalErr.StatusCode()).JSON(internalErr)
}

func FiberLimiter() fiber.Handler {
	maxLimiter, err := strconv.Atoi(config.Environment.RatelimiterPerMinute)
	if err != nil {
		logger.Warn("FIBER_CONFIG_LIMITER", err)
	}

	return limiter.New(limiter.Config{
		Max:        maxLimiter,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			err := restErrors.NewTooManyRequestsError("too many requests")
			return c.Status(fiber.StatusTooManyRequests).JSON(shared.NewResponse(err))
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
		LimiterMiddleware:      limiter.FixedWindow{},
	})
}
