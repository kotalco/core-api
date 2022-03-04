package config

import (
	"github.com/gofiber/fiber/v2"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/api/pkg/logger"

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
	go logger.Panic("PANICKING", err)

	internalErr := restErrors.NewInternalServerError("some thing went wrong...")

	return c.Status(internalErr.Status).JSON(internalErr)
}
