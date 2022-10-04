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
	"github.com/kotalco/cloud-api/pkg/k8s"
	"github.com/kotalco/cloud-api/pkg/security"
	subscriptionAPI "github.com/kotalco/cloud-api/pkg/subscription"
	"net/http"
	"time"
)

const KUBE_SYSTEM_NAMESPACE = "kube-system"

var (
	subscriptionAPIService = subscriptionAPI.NewSubscriptionService()
	ecService              = security.NewEllipticCurve()
	namespaceService       = k8s.NewNamespaceService()
)

//Acknowledgement accept the user activation_key
//calls the subscription_api to get the user license details and signature (which have been created with ecc using the system acc private key)
//the system uses its public key to validate the integrity of the subscription details, to the signature
//activate the user according to his subscription plan
//Sync the subscription back to the subscription-platform with the clusterId
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

	validSub := subscriptionAPI.IsValid()
	if !validSub {
		expiredRestErr := restErrors.RestErr{
			Status:  http.StatusGone,
			Message: "invalid subscription",
			Name:    "STATUS_GONE",
		}
		return c.Status(expiredRestErr.Status).JSON(expiredRestErr)
	}

	//get clusterID in the form of kube-system namespaceID
	//since the cluster has no ID we alias the clusterId with kube-system namespace ID coz its immutable and unique
	ns, err := namespaceService.Get(KUBE_SYSTEM_NAMESPACE)
	if err != nil {
		subscriptionAPI.Reset()
		err.Message = "couldn't sync subscription to Kotal"
		return c.Status(err.Status).JSON(err)
	}

	err = subscriptionAPIService.SyncAcknowledgment(dto.ActivationKey, string(ns.UID))
	if err != nil {
		subscriptionAPI.Reset()
		err.Message = "couldn't sync subscription to Kotal"
		return c.Status(err.Status).JSON(err)
	}
	//Todo validate sync-acknowledgment response with ecc
	return c.Status(http.StatusOK).JSON(shared.NewResponse(shared.SuccessMessage{Message: "subscription activated"}))
}

func Current(c *fiber.Ctx) error {
	if subscriptionAPI.SubscriptionDetails == nil {
		go logger.Error("CURRENT_SUBSCRIPTION", errors.New("user have no active subscription and passed the middleware check"))
		expiredRestErr := restErrors.RestErr{
			Status:  http.StatusGone,
			Message: "invalid subscription",
			Name:    "STATUS_GONE",
		}
		return c.Status(expiredRestErr.Status).JSON(expiredRestErr)
	}

	return c.Status(http.StatusOK).JSON(shared.NewResponse(subscriptionAPI.SubscriptionDetails))
}
