package health

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/subscriptions-api/pkg/config"
	httpCheck "github.com/kotalco/subscriptions-api/pkg/health/http"
	psqlCheck "github.com/kotalco/subscriptions-api/pkg/health/psql"
	stripeCheck "github.com/kotalco/subscriptions-api/pkg/health/stripe"
	"net/http"
)

func Healthz(c *fiber.Ctx) error {
	h := New()
	err := h.Register(Config{
		Name:      "PSQL",
		SkipOnErr: false,
		Measure:   psqlCheck.New(psqlCheck.Config{DBServerURL: config.Environment.DatabaseServerURL}),
	}, Config{
		Name:      "Stripe",
		SkipOnErr: false,
		Measure:   stripeCheck.New(stripeCheck.Config{StripeAPIKey: config.Environment.StripeAPIKey}),
	}, Config{
		Name:      "Google",
		SkipOnErr: false,
		Measure: httpCheck.New(httpCheck.Config{
			URL:            "https://google.com",
			RequestTimeout: 1,
		}),
	})

	if err != nil {
		healthDto := ResponseDto{
			Checks: []Check{},
			Status: StatusUnavailable,
		}
		return c.Status(http.StatusInternalServerError).JSON(healthDto)
	}

	res := h.Measure(c)

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
