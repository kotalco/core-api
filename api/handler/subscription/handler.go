package subscription

import (
	"errors"
	"github.com/gofiber/fiber/v2"
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
)

//Acknowledgement accept the user activation_key
// Runs subscription acknowledgement
//validate subscription
func Acknowledgement(c *fiber.Ctx) error {
	//accept and validate the activation key
	dto := new(subscription.AcknowledgementRequestDto)
	if intErr := c.BodyParser(dto); intErr != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.Status).JSON(badReq)
	}

	err := subscription.Validate(dto)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	err = subscriptionService.Acknowledgment(dto.ActivationKey)
	if err != nil {
		subscriptionAPI.Reset()
		return c.Status(err.Status).JSON(err)
	}

	validSub := subscriptionAPI.IsValid()
	if !validSub {
		err = &restErrors.RestErr{
			Status:  http.StatusForbidden,
			Message: "invalid subscription",
			Name:    middleware.InvalidSubscriptionStatusMessage,
		}
		return c.Status(err.Status).JSON(err)
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
