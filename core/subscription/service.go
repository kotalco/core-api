package subscription

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/kotalco/cloud-api/pkg/config"
	"github.com/kotalco/cloud-api/pkg/k8s"
	"github.com/kotalco/cloud-api/pkg/security"
	subscriptionAPI "github.com/kotalco/cloud-api/pkg/subscription"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
)

const KUBE_SYSTEM_NAMESPACE = "kube-system"

var (
	subscriptionAPIService = subscriptionAPI.NewSubscriptionService()
	ecService              = security.NewEllipticCurve()
	namespaceService       = k8s.NewNamespaceService()
)

type service struct{}

type IService interface {
	//Acknowledgment accepts user activation key
	//gets kube-system namespace id and since it's immutable and unique the system would use it as the clusterID
	//calls the subscription_api to get the user license details and signature (which have been created with ecc using the system acc private key)
	//the system uses its public key to validate the integrity of the subscription details, to the signature
	//activate the user according to his subscription plan
	Acknowledgment(activationKey string) restErrors.IRestErr
	//CurrentTimestamp returns the current timestamp
	//by calling the subscription platform and validating the signature of this timestamp using ecc
	CurrentTimestamp() (int64, restErrors.IRestErr)
}

func NewService() IService {
	return &service{}
}

func (subService *service) Acknowledgment(activationKey string) restErrors.IRestErr {
	//get clusterID in the form of kube-system namespaceID
	//since the cluster has no ID we alias the clusterId with kube-system namespace ID coz its immutable and unique
	ns, err := namespaceService.Get(KUBE_SYSTEM_NAMESPACE)
	if err != nil {
		subscriptionAPI.Reset()
		err = restErrors.NewBadRequestError("can't get cluster details")
		return err
	}

	responseData, err := subscriptionAPIService.Acknowledgment(activationKey, string(ns.UID))
	if err != nil {
		return err
	}

	var responseBody map[string]LicenseAcknowledgmentDto
	intErr := json.Unmarshal(responseData, &responseBody)
	if intErr != nil {
		go logger.Error(subService.Acknowledgment, intErr)
		err = restErrors.NewInternalServerError("can't activate subscription")
		return err
	}
	licenseAcknowledgmentDto := responseBody["data"]
	//validate the signature
	decodedPub, intErr := ecService.DecodePublic(config.Environment.ECCPublicKey)
	if intErr != nil {
		go logger.Error(subService.Acknowledgment, intErr)
		err = restErrors.NewInternalServerError("can't activate subscription")
		return err
	}

	subscriptionBytes, intErr := json.Marshal(licenseAcknowledgmentDto.Subscription)
	if intErr != nil {
		go logger.Error(subService.Acknowledgment, intErr)
		err = restErrors.NewInternalServerError("can't activate subscription")
		return err
	}

	signatureBytes, intErr := base64.StdEncoding.DecodeString(licenseAcknowledgmentDto.Signature)
	if intErr != nil {
		go logger.Error(subService.Acknowledgment, intErr)
		err = restErrors.NewInternalServerError("can't activate subscription")
		return err
	}

	valid, intErr := ecService.VerifySignature(subscriptionBytes, signatureBytes, decodedPub)
	if intErr != nil {
		go logger.Error(subService.Acknowledgment, intErr)
		err = restErrors.NewInternalServerError("can't activate subscription")
		return err
	}
	if !valid {
		go logger.Error(subService.Acknowledgment, errors.New("invalid signature"))
		err = restErrors.NewInternalServerError("can't activate subscription")
		return err
	}

	//get last time
	currentTime, err := subService.CurrentTimestamp()
	if err != nil {
		go logger.Error(subService.CurrentTimestamp, err)
		return err
	}

	//save last check data
	subscriptionAPI.CheckDate = currentTime

	//assign license details
	subscriptionAPI.SubscriptionDetails = &subscriptionAPI.SubscriptionDetailsDto{
		Status:       licenseAcknowledgmentDto.Subscription.Status,
		StartDate:    licenseAcknowledgmentDto.Subscription.StartDate,
		EndDate:      licenseAcknowledgmentDto.Subscription.EndDate,
		Name:         licenseAcknowledgmentDto.Subscription.Name,
		CanceledAt:   licenseAcknowledgmentDto.Subscription.CanceledAt,
		TrialStartAt: licenseAcknowledgmentDto.Subscription.TrialStartAt,
		TrialEndAt:   licenseAcknowledgmentDto.Subscription.EndDate,
		NodesLimit:   licenseAcknowledgmentDto.Subscription.NodesLimit,
	}

	return nil
}

func (subService *service) CurrentTimestamp() (int64, restErrors.IRestErr) {
	responseData, err := subscriptionAPIService.CurrentTimeStamp()
	if err != nil {
		return 0, err
	}

	var responseBody map[string]CurrentTimeStampDto
	intErr := json.Unmarshal(responseData, &responseBody)
	if intErr != nil {
		go logger.Error(subService.CurrentTimestamp, intErr)
		err = restErrors.NewInternalServerError("something went wrong")
		return 0, err
	}
	currentTimestampDto := responseBody["data"]
	//validate the signature
	decodedPub, intErr := ecService.DecodePublic(config.Environment.ECCPublicKey)
	if intErr != nil {
		go logger.Error(subService.CurrentTimestamp, intErr)
		err = restErrors.NewInternalServerError("something went wrong")
		return 0, err
	}

	currentTimestampDtoBytes, intErr := json.Marshal(currentTimestampDto.Time)
	if intErr != nil {
		go logger.Error(subService.CurrentTimestamp, intErr)
		err = restErrors.NewInternalServerError("something went wrong")
		return 0, err
	}

	signatureBytes, intErr := base64.StdEncoding.DecodeString(currentTimestampDto.Signature)
	if intErr != nil {
		go logger.Error(subService.CurrentTimestamp, intErr)
		err = restErrors.NewInternalServerError("something went wrong")
		return 0, err
	}

	valid, intErr := ecService.VerifySignature(currentTimestampDtoBytes, signatureBytes, decodedPub)
	if intErr != nil {
		go logger.Error(subService.CurrentTimestamp, intErr)
		err = restErrors.NewInternalServerError("something went wrong")
		return 0, err
	}
	if !valid {
		go logger.Error(subService.CurrentTimestamp, errors.New("invalid signature"))
		err = restErrors.NewInternalServerError("something went wrong")
		return 0, err
	}

	return currentTimestampDto.Time.CurrentTime, err
}