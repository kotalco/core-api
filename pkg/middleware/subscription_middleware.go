package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/internal/subscription"
	subscriptionAPI "github.com/kotalco/cloud-api/pkg/subscription"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	"net/http"
	"time"
)

const (
	InvalidSubscriptionStatusMessage = "INVALID_SUBSCRIPTION"
)

var (
	subscriptionService = subscription.NewService()
)

func IsSubscription(c *fiber.Ctx) error {
	//get last time
	currenTimeInUnix, err := subscriptionService.CurrentTimestamp()
	if err != nil {
		go logger.Error("IS_SUBSCRIPTION_MIDDLEWARE", err)
		return err
	}

	lastCheckDate := time.Unix(subscriptionAPI.CheckDate, 0)
	lastCheckDateInUnixWithGracePeriod := lastCheckDate.Add(time.Hour * 24).Unix()

	if lastCheckDateInUnixWithGracePeriod < currenTimeInUnix {
		//check if activation key exits
		if subscriptionAPI.ActivationKey == "" {
			invalidSubErr := restErrors.RestErr{Status: http.StatusForbidden, Message: "invalid subscription", Name: InvalidSubscriptionStatusMessage}
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
			Status:  http.StatusForbidden,
			Message: "invalid subscription",
			Name:    InvalidSubscriptionStatusMessage,
		}
		return c.Status(invalidSubErr.Status).JSON(invalidSubErr)
	}
	c.Locals("subscriptionDetails", *subscriptionAPI.SubscriptionDetails)

	c.Next()
	return nil
}
