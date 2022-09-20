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
	elapsedTime := time.Now().Unix() - subscriptionAPI.CheckDate
	if elapsedTime > int64(time.Hour)*24 {
		//check if activation key exits
		if subscriptionAPI.ActivationKey == "" {
			invalidSubErr := restErrors.RestErr{Status: http.StatusGone, Message: "invalid subscription", Name: "STATUS_GONE"}
			return c.Status(invalidSubErr.Status).JSON(invalidSubErr)
		}
		//call subscription platform for acknowledgement
		responseData, err := subscriptionAPIService.Acknowledgment(subscriptionAPI.ActivationKey)
		if err != nil {
			return c.Status(err.Status).JSON(err)
		}

		var responseBody map[string]subscription.LicenseAcknowledgmentDto
		intErr := json.Unmarshal(responseData, &responseBody)
		if intErr != nil {
			go logger.Error("ACKNOWLEDGEMENT_HANDLER", intErr)
			invalidSubErr := restErrors.RestErr{Status: http.StatusGone, Message: "invalid subscription", Name: "STATUS_GONE"}
			return c.Status(invalidSubErr.Status).JSON(invalidSubErr)
		}
		licenseAcknowledgmentDto := responseBody["data"]

		//validate the signature
		decodedPub, intErr := ecService.DecodePublic(config.EnvironmentConf["ECC_PUBLIC_KEY"])
		if intErr != nil {
			go logger.Error("ACKNOWLEDGEMENT_HANDLER", intErr)
			invalidSubErr := restErrors.RestErr{Status: http.StatusGone, Message: "invalid subscription", Name: "STATUS_GONE"}
			return c.Status(invalidSubErr.Status).JSON(invalidSubErr)
		}

		subscriptionBytes, intErr := json.Marshal(licenseAcknowledgmentDto.Subscription)
		if intErr != nil {
			go logger.Error("ACKNOWLEDGEMENT_HANDLER", intErr)
			invalidSubErr := restErrors.RestErr{Status: http.StatusGone, Message: "invalid subscription", Name: "STATUS_GONE"}
			return c.Status(invalidSubErr.Status).JSON(invalidSubErr)
		}

		signatureBytes, intErr := base64.StdEncoding.DecodeString(licenseAcknowledgmentDto.Signature)
		if intErr != nil {
			go logger.Error("ACKNOWLEDGEMENT_HANDLER", intErr)
			invalidSubErr := restErrors.RestErr{Status: http.StatusGone, Message: "invalid subscription", Name: "STATUS_GONE"}
			return c.Status(invalidSubErr.Status).JSON(invalidSubErr)
		}

		valid, intErr := ecService.VerifySignature(subscriptionBytes, signatureBytes, decodedPub)
		if intErr != nil {
			go logger.Error("ACKNOWLEDGEMENT_HANDLER", intErr)
			invalidSubErr := restErrors.RestErr{Status: http.StatusGone, Message: "invalid subscription", Name: "STATUS_GONE"}
			return c.Status(invalidSubErr.Status).JSON(invalidSubErr)
		}
		if !valid {
			go logger.Error("ACKNOWLEDGEMENT_HANDLER", errors.New("invalid signature"))
			invalidSubErr := restErrors.RestErr{Status: http.StatusGone, Message: "invalid subscription", Name: "STATUS_GONE"}
			return c.Status(invalidSubErr.Status).JSON(invalidSubErr)
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
