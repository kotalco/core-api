package http

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/subscriptions-api/pkg/logger"
	"net/http"
	"time"
)

const defaultRequestTimeout = 5 * time.Second

type Config struct {
	URL string
	// If not set - 5 seconds
	RequestTimeout time.Duration
}

func New(config Config) func(ctx *fiber.Ctx) error {
	if config.RequestTimeout == 0 {
		config.RequestTimeout = defaultRequestTimeout
	}

	return func(ctx *fiber.Ctx) error {

		req, err := http.NewRequest(http.MethodGet, config.URL, bytes.NewReader([]byte{}))
		if err != nil {
			go logger.Error("HTTP_HEALTH_CHECK", err)
			return fmt.Errorf("creating the request for the health check failed: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Connection", "close")

		client := http.Client{
			Timeout: defaultRequestTimeout,
		}
		res, err := client.Do(req)
		if err != nil {
			go logger.Error("HTTP_HEALTH_CHECK", err)
			return fmt.Errorf("making the request for the health check failed: %w", err)
		}

		defer res.Body.Close()

		if res.StatusCode >= http.StatusInternalServerError {
			go logger.Error("HTTP_HEALTH_CHECK", errors.New("remote service is not available at the moment"))
			return errors.New("remote service is not available at the moment")
		}

		return nil
	}
}
