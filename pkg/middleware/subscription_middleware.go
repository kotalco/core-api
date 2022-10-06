package middleware

import (
	"github.com/gofiber/fiber/v2"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/cloud-api/internal/subscription"
	subscriptionAPI "github.com/kotalco/cloud-api/pkg/subscription"
	"net/http"
	"time"
)

var (
	subscriptionService = subscription.NewService()
)

func IsSubscription(c *fiber.Ctx) error {
	elapsedTime := time.Now().Unix() - subscriptionAPI.CheckDate
	if elapsedTime > int64(time.Hour)*24 {
		//check if activation key exits
		if subscriptionAPI.ActivationKey == "" {
			invalidSubErr := restErrors.RestErr{Status: http.StatusGone, Message: "invalid subscription", Name: "STATUS_GONE"}
			return c.Status(invalidSubErr.Status).JSON(invalidSubErr)
		}

		//run subscription validity check coz the elapsed time exceeded check time
		err := subscriptionService.Acknowledgment(subscriptionAPI.ActivationKey)
		if err != nil {
			return c.Status(err.Status).JSON(err)
		}
	}

	// validate if subscription valid
	validSub := subscriptionAPI.IsValid()
	if !validSub {
		invalidSubErr := restErrors.RestErr{
			Status:  http.StatusGone,
			Message: "invalid subscription",
			Name:    "STATUS_GONE",
		}
		return c.Status(invalidSubErr.Status).JSON(invalidSubErr)
	}
	c.Locals("subscriptionDetails", *subscriptionAPI.SubscriptionDetails)

	c.Next()
	return nil
}
