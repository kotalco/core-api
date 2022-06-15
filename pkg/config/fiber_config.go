package config

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/api/pkg/shared"

	"strconv"
	"time"
)

func FiberConfig() fiber.Config {
	readTimeoutSecondsCount, _ := strconv.Atoi(EnvironmentConf["SERVER_READ_TIMEOUT"])
	return fiber.Config{
		ReadTimeout:  time.Second * time.Duration(readTimeoutSecondsCount),
		ErrorHandler: defaultErrorHandler,
	}
}

//defaultErrorHandler used to catch all unhandled  run time resterror mainly panics
//logs resterror using logger pkg
//return custom error struct using restError pkg
var defaultErrorHandler = func(c *fiber.Ctx, err error) error {
	go logger.Panic("PANICKING_UNHANDLED_ERROR", err)
	internalErr := restErrors.NewInternalServerError("some thing went wrong")

	return c.Status(internalErr.Status).JSON(internalErr)
}

func FiberLimiter() fiber.Handler {
	maxLimiter, err := strconv.Atoi(EnvironmentConf["RATE_LIMITER_PER_MINUTE"])
	if err != nil {
		logger.Panic("FIBER_CONFIG_LIMITER", err)
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
