package subscription

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/api/pkg/shared"
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

//Acknowledgement accept the user activation_key
//calls the subscription_api to get the user license details and signature (which have been created with ecc using the system acc private key)
//the system uses its public key to validate the integrity of the subscription details, to the signature
//activate the user according to his subscription plan
func Acknowledgement(c *fiber.Ctx) error {
	//accept and validate the activation key
	dto := new(subscription.AcknowledgementRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.Status).JSON(badReq)
	}

	restErr := subscription.Validate(dto)
	if restErr != nil {
		return c.Status(restErr.Status).JSON(restErr)
	}

	responseData, err := subscriptionAPIService.Acknowledgment(dto.ActivationKey)
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
		err = restErrors.NewInternalServerError("can't activate subscription")
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
	subscriptionAPI.CheckDate = time.Now().Unix()

	//assign license details
	subscriptionAPI.SubscriptionDetails = &subscriptionAPI.SubscriptionDetailsDto{
		Status:     licenseAcknowledgmentDto.Subscription.Status,
		StartDate:  licenseAcknowledgmentDto.Subscription.StartDate,
		EndDate:    licenseAcknowledgmentDto.Subscription.EndDate,
		Name:       licenseAcknowledgmentDto.Subscription.Name,
		CanceledAt: licenseAcknowledgmentDto.Subscription.CanceledAt,
	}

	//store activation key
	subscriptionAPI.ActivationKey = dto.ActivationKey

	return c.Status(http.StatusOK).JSON(shared.NewResponse(shared.SuccessMessage{Message: "subscription activated"}))
}
