package health

import (
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

type IHealth interface {
	Measure(list ...Config) *ResponseDto
}
type Health struct {
}

// Config carries the parameters to run the check.
type Config struct {
	// Name is the name of the resource to be checked.
	Name string
	// SkipOnErr if set to true, it will retrieve StatusOK providing the error message from the failed resource.
	SkipOnErr bool
	// Check is the func which executes the check.
	Measure func() error
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
func New() IHealth {
	return &Health{}
}

// Measure runs all the registered health checks and returns summary status
func (h *Health) Measure(list ...Config) *ResponseDto {
	res := new(ResponseDto)
	res.Status = StatusOK
	checks := make([]Check, 0)

	//run the checks
	for _, v := range list {

		check := &Check{
			Name:      v.Name,
			Status:    StatusOK,
			Timestamp: time.Now(),
		}
		err := v.Measure()
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
