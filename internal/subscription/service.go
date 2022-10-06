package subscription

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/cloud-api/pkg/config"
	"github.com/kotalco/cloud-api/pkg/k8s"
	"github.com/kotalco/cloud-api/pkg/security"
	subscriptionAPI "github.com/kotalco/cloud-api/pkg/subscription"
	"time"
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
	Acknowledgment(activationKey string) *restErrors.RestErr
}

func NewService() IService {
	return &service{}
}

func (subService *service) Acknowledgment(activationKey string) *restErrors.RestErr {
	//get clusterID in the form of kube-system namespaceID
	//since the cluster has no ID we alias the clusterId with kube-system namespace ID coz its immutable and unique
	ns, err := namespaceService.Get(KUBE_SYSTEM_NAMESPACE)
	if err != nil {
		subscriptionAPI.Reset()
		err.Message = "can't get cluster details"
		return err
	}

	responseData, err := subscriptionAPIService.Acknowledgment(activationKey, string(ns.UID))
	if err != nil {
		return err
	}

	var responseBody map[string]LicenseAcknowledgmentDto
	intErr := json.Unmarshal(responseData, &responseBody)
	if intErr != nil {
		go logger.Error("ACKNOWLEDGEMENT_HANDLER", intErr)
		err = restErrors.NewInternalServerError("can't activate subscription")
		return err
	}
	licenseAcknowledgmentDto := responseBody["data"]
	//validate the signature
	decodedPub, intErr := ecService.DecodePublic(config.EnvironmentConf["ECC_PUBLIC_KEY"])
	if intErr != nil {
		go logger.Error("ACKNOWLEDGEMENT_HANDLER", intErr)
		err = restErrors.NewInternalServerError("can't activate subscription")
		return err
	}

	subscriptionBytes, intErr := json.Marshal(licenseAcknowledgmentDto.Subscription)
	if intErr != nil {
		go logger.Error("ACKNOWLEDGEMENT_HANDLER", intErr)
		err = restErrors.NewInternalServerError("can't activate subscription")
		return err
	}

	signatureBytes, intErr := base64.StdEncoding.DecodeString(licenseAcknowledgmentDto.Signature)
	if intErr != nil {
		go logger.Error("ACKNOWLEDGEMENT_HANDLER", intErr)
		err = restErrors.NewInternalServerError("can't activate subscription")
		return err
	}

	valid, intErr := ecService.VerifySignature(subscriptionBytes, signatureBytes, decodedPub)
	if intErr != nil {
		go logger.Error("ACKNOWLEDGEMENT_HANDLER", intErr)
		err = restErrors.NewInternalServerError("can't activate subscription")
		return err
	}
	if !valid {
		go logger.Error("ACKNOWLEDGEMENT_HANDLER", errors.New("invalid signature"))
		err = restErrors.NewInternalServerError("can't activate subscription")
		return err
	}

	//save last check data
	//Todo time externally request
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
	subscriptionAPI.ActivationKey = activationKey

	return nil
}