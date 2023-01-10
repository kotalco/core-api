package stripe

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/subscriptions-api/pkg/logger"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/client"
)

type Config struct {
	StripeAPIKey string
}

func New(config Config) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) (checkErr error) {
		stripe.Key = config.StripeAPIKey
		sc := &client.API{}
		sc.Init(config.StripeAPIKey, nil)
		iter := sc.Products.List(&stripe.ProductListParams{})
		if iter.Err() != nil {
			go logger.Error("STRIPE_HEALTH_CHECK", iter.Err())
			return fmt.Errorf("stripe health check failed on connect: %w", iter.Err())
		}
		return
	}
}
