package health

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/pkg/config"
	httpCheck "github.com/kotalco/cloud-api/pkg/health/http"
	psqlCheck "github.com/kotalco/cloud-api/pkg/health/psql"
	"github.com/kotalco/cloud-api/pkg/subscription"
	"net/http"
)

var h IHealth
var newHealthCheckService = New

func Healthz(c *fiber.Ctx) error {
	h = newHealthCheckService()
	err := h.Register(Config{
		Name:      "PSQL",
		SkipOnErr: false,
		Measure:   psqlCheck.New(psqlCheck.Config{DBServerURL: config.Environment.DatabaseServerURL}),
	},
		Config{
			Name:      "Subscription-Api",
			SkipOnErr: false,
			Measure: httpCheck.New(httpCheck.Config{
				URL:            config.Environment.SubscriptionAPIBaseURL + subscription.CURRENT_TIMESTAMP,
				RequestTimeout: 5,
			}),
		},
	)

	if err != nil {
		healthDto := ResponseDto{
			Checks: []Check{},
			Status: StatusUnavailable,
		}
		return c.Status(http.StatusServiceUnavailable).JSON(healthDto)
	}

	res := h.Measure()

	return c.Status(intStatus(res.Status)).JSON(res)

}

func intStatus(status Status) int {
	switch status {
	case StatusOK:
		return http.StatusOK
	case StatusPartiallyAvailable:
		return http.StatusOK
	case StatusUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusOK
	}
}
