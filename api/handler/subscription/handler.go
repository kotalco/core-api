package subscription

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/internal/setting"
	"github.com/kotalco/cloud-api/internal/subscription"
	"github.com/kotalco/cloud-api/pkg/middleware"
	subscriptionAPI "github.com/kotalco/cloud-api/pkg/subscription"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	"github.com/kotalco/community-api/pkg/shared"
	"net/http"
)

var (
	subscriptionService = subscription.NewService()
	settingService      = setting.NewService()
)

// Acknowledgement accept the user activation_key
// Runs subscription acknowledgement
// validate subscription
func Acknowledgement(c *fiber.Ctx) error {
	//accept and validate the activation key
	dto := new(subscription.AcknowledgementRequestDto)
	if intErr := c.BodyParser(dto); intErr != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}

	err := subscription.Validate(dto)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	err = subscriptionService.Acknowledgment(dto.ActivationKey)
	if err != nil {
		subscriptionAPI.Reset()
		return c.Status(err.StatusCode()).JSON(err)
	}

	validSub := subscriptionAPI.IsValid()
	if !validSub {
		err = &restErrors.RestErr{
			Status:  http.StatusForbidden,
			Message: "invalid subscription",
			Name:    middleware.InvalidSubscriptionStatusMessage,
		}
		return c.Status(err.StatusCode()).JSON(err)
	}

	//store subscription activation key
	err = settingService.WithoutTransaction().ConfigureActivationKey(dto.ActivationKey)
	if err != nil {
		//reset the subscription status can't coz we can't save the activation key
		subscriptionAPI.Reset()
		go logger.Warn("Acknowledgement", err)
		intErr := restErrors.NewInternalServerError("can't save activation key")
		return c.Status(intErr.StatusCode()).JSON(intErr)
	}

	return c.Status(http.StatusOK).JSON(shared.NewResponse(shared.SuccessMessage{Message: "subscription activated"}))
}

func Current(c *fiber.Ctx) error {
	if subscriptionAPI.SubscriptionDetails == nil {
		go logger.Error("CURRENT_SUBSCRIPTION", errors.New("user have no active subscription and passed the middleware check"))
		expiredRestErr := restErrors.RestErr{
			Status:  http.StatusForbidden,
			Message: "invalid subscription",
			Name:    middleware.InvalidSubscriptionStatusMessage,
		}
		return c.Status(expiredRestErr.Status).JSON(expiredRestErr)
	}

	return c.Status(http.StatusOK).JSON(shared.NewResponse(subscriptionAPI.SubscriptionDetails))
}
