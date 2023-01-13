package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/internal/setting"
	"github.com/kotalco/cloud-api/internal/subscription"
	"github.com/kotalco/cloud-api/pkg/k8s/statefulset"
	subscriptionAPI "github.com/kotalco/cloud-api/pkg/subscription"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	"net/http"
	"time"
)

const (
	InvalidSubscriptionStatusMessage = "INVALID_SUBSCRIPTION"
	NodeLimitStatusMessage           = "NODES_LIMIT"
)

var (
	subscriptionService = subscription.NewService()
	statefulSetService  = statefulset.NewService()
	settingService      = setting.NewService()
)

func IsSubscription(c *fiber.Ctx) error {
	//get last time
	currenTimeInUnix, err := subscriptionService.CurrentTimestamp()
	if err != nil {
		go logger.Error("IS_SUBSCRIPTION_MIDDLEWARE", err)
		return err
	}

	//get activation key
	activationKey, err := settingService.WithoutTransaction().GetActivationKey()
	if err != nil {
		go logger.Warn("IS_SUBSCRIPTION_MIDDLEWARE", err)
		invalidSubErr := restErrors.RestErr{Status: http.StatusForbidden, Message: "invalid subscription", Name: InvalidSubscriptionStatusMessage}
		return c.Status(invalidSubErr.Status).JSON(invalidSubErr)
	}

	lastCheckDate := time.Unix(subscriptionAPI.CheckDate, 0)
	lastCheckDateInUnixWithGracePeriod := lastCheckDate.Add(time.Hour * 1).Unix()

	//do periodic check with every login attempt , and every hour
	if lastCheckDateInUnixWithGracePeriod < currenTimeInUnix || c.Path() == "/api/v1/sessions" {
		//check if activation key exits
		if activationKey == "" {
			invalidSubErr := restErrors.RestErr{Status: http.StatusForbidden, Message: "invalid subscription", Name: InvalidSubscriptionStatusMessage}
			return c.Status(invalidSubErr.Status).JSON(invalidSubErr)
		}

		//run subscription validity check coz the elapsed time exceeded check time
		err := subscriptionService.Acknowledgment(activationKey)
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

	return c.Next()
}

func NodesLimitProtected(c *fiber.Ctx) error {
	//validate nodes limit
	limit, err := statefulSetService.Count()
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	if limit >= subscriptionAPI.SubscriptionDetails.NodesLimit {
		err := restErrors.RestErr{
			Message: "reached nodes limit",
			Status:  http.StatusForbidden,
			Name:    NodeLimitStatusMessage,
		}
		return c.Status(err.Status).JSON(err)
	}

	return c.Next()
}
