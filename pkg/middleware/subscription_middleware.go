package middleware

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/cloud-api/internal/subscription"
	"github.com/kotalco/cloud-api/pkg/config"
	"github.com/kotalco/cloud-api/pkg/security"
	subscriptionAPI "github.com/kotalco/cloud-api/pkg/subscription"
	"net/http"
	"time"
)

var (
	subscriptionAPIService = subscriptionAPI.NewSubscriptionService()
	ecService              = security.NewEllipticCurve()
)

func IsSubscription(c *fiber.Ctx) error {
	if subscriptionAPI.CheckDate < time.Now().Add(-time.Hour*24).Unix() {
		responseData, err := subscriptionAPIService.Acknowledgment(subscriptionAPI.ActivationKey)
		if err != nil {
			return c.Status(err.Status).JSON(err)
		}

		var responseBody map[string]subscription.LicenseAcknowledgmentDto
		intErr := json.Unmarshal(responseData, &responseBody)
		if intErr != nil {
			go logger.Error("ACKNOWLEDGEMENT_HANDLER", intErr)
			err = restErrors.NewInternalServerError("can't activate subscription")
			return c.Status(err.Status).JSON(err)
		}
		licenseAcknowledgmentDto := responseBody["data"]

		//validate the signature
		decodedPub, intErr := ecService.DecodePublic(config.EnvironmentConf["ECC_PUBLIC_KEY"])
		if intErr != nil {
			go logger.Error("ACKNOWLEDGEMENT_HANDLER", intErr)
			err := restErrors.NewInternalServerError("can't activate subscription")
			return c.Status(err.Status).JSON(err)
		}

		subscriptionBytes, intErr := json.Marshal(licenseAcknowledgmentDto.Subscription)
		if intErr != nil {
			go logger.Error("ACKNOWLEDGEMENT_HANDLER", intErr)
			err = restErrors.NewInternalServerError("can't activate subscription")
			return c.Status(err.Status).JSON(err)
		}

		signatureBytes, intErr := base64.StdEncoding.DecodeString(licenseAcknowledgmentDto.Signature)
		if intErr != nil {
			go logger.Error("ACKNOWLEDGEMENT_HANDLER", intErr)
			err = restErrors.NewInternalServerError("can't activate subscription")
			return c.Status(err.Status).JSON(err)
		}

		valid, intErr := ecService.VerifySignature(subscriptionBytes, signatureBytes, decodedPub)
		if intErr != nil {
			go logger.Error("ACKNOWLEDGEMENT_HANDLER", intErr)
			err = restErrors.NewInternalServerError("can't activate subscription")
			return c.Status(err.Status).JSON(err)
		}
		if !valid {
			go logger.Error("ACKNOWLEDGEMENT_HANDLER", errors.New("invalid signature"))
			err = restErrors.NewInternalServerError("can't activate subscription")
			return c.Status(err.Status).JSON(err)
		}

		//save last check data
		//Todo time externally request
		subscriptionAPI.CheckDate = time.Now().Unix()

		//assign license details
		subscriptionAPI.SubscriptionDetails.Status = licenseAcknowledgmentDto.Subscription.Status
		subscriptionAPI.SubscriptionDetails.StartDate = licenseAcknowledgmentDto.Subscription.StartDate
		subscriptionAPI.SubscriptionDetails.EndDate = licenseAcknowledgmentDto.Subscription.EndDate
		subscriptionAPI.SubscriptionDetails.Name = licenseAcknowledgmentDto.Subscription.Name
		subscriptionAPI.SubscriptionDetails.CanceledAt = licenseAcknowledgmentDto.Subscription.CanceledAt
	}

	validSub := subscriptionAPI.IsValid()
	if !validSub {
		expiredRestErr := restErrors.RestErr{
			Status:  http.StatusGone,
			Message: "invalid subscription",
		}
		return c.Status(expiredRestErr.Status).JSON(expiredRestErr)
	}
	c.Locals("subscriptionDetails", *subscriptionAPI.SubscriptionDetails)

	c.Next()
	return nil
}
