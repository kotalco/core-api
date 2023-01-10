package health

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"time"
)

// Status type represents health status
type Status string

// Possible health statuses
const (
	StatusOK                 Status = "OK"
	StatusPartiallyAvailable Status = "Partially Available"
	StatusUnavailable        Status = "Unavailable"
)

type Health struct {
	configs map[string]Config
}

// Config carries the parameters to run the check.
type Config struct {
	// Name is the name of the resource to be checked.
	Name string
	// SkipOnErr if set to true, it will retrieve StatusOK providing the error message from the failed resource.
	SkipOnErr bool
	// Check is the func which executes the check.
	Measure func(ctx *fiber.Ctx) error
}

// Check represents the health check response.
type Check struct {
	Name string `json:"name"`
	// Status is the check status.
	Status Status `json:"status"`
	// Timestamp is the time in which the check occurred.
	Timestamp time.Time `json:"timestamp"`
	// Failure holds the failed check error messages.
	Failure string `json:"failure,omitempty"`
}

// ResponseDto HealthCheckResponseDto represents the health check response for all services
type ResponseDto struct {
	Checks []Check `json:"checks"`
	Status Status  `json:"status"`
}

// New instantiates and build new health check container
func New() *Health {
	return &Health{
		configs: make(map[string]Config),
	}
}

// Register registers a check config to be performed.
func (h *Health) Register(list ...Config) error {
	for _, c := range list {
		if c.Name == "" {
			return errors.New("health check must have a name to be registered")
		}

		if _, ok := h.configs[c.Name]; ok {
			return fmt.Errorf("health check %q is already registered", c.Name)
		}
		h.configs[c.Name] = c
	}

	return nil
}

// Measure runs all the registered health checks and returns summary status
func (h *Health) Measure(ctx *fiber.Ctx) *ResponseDto {
	res := new(ResponseDto)
	res.Status = StatusOK
	checks := make([]Check, 0)

	//run the checks
	for k, v := range h.configs {

		check := &Check{
			Name:      k,
			Status:    StatusOK,
			Timestamp: time.Now(),
		}
		err := v.Measure(ctx)
		if err != nil {
			check.Status = StatusUnavailable
			check.Failure = err.Error()
			res.Status = getAvailability(res.Status, v.SkipOnErr)
		}
		checks = append(checks, *check)
	}
	res.Checks = checks
	return res

}

// getAvailability return the for the whole health check status
func getAvailability(responseStatus Status, skipOnErr bool) Status {
	if responseStatus == StatusPartiallyAvailable {
		return StatusPartiallyAvailable
	}
	if skipOnErr {
		return StatusPartiallyAvailable
	}
	return StatusUnavailable
}
