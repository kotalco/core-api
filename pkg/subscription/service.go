package subscription

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/kotalco/cloud-api/pkg/config"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	ACKNOWLEDGEMENT   = "/api/v1/license/acknowledgment"
	CURRENT_TIMESTAMP = "/api/v1/timestamp/current"
)

type ISubscriptionService interface {
	Acknowledgment(activationKey string, clusterID string) ([]byte, *restErrors.RestErr)
	CurrentTimeStamp() ([]byte, *restErrors.RestErr)
}

type subscriptionService struct{}

func NewSubscriptionService() ISubscriptionService {
	return &subscriptionService{}
}

func (subApi *subscriptionService) Acknowledgment(activationKey string, clusterID string) ([]byte, *restErrors.RestErr) {
	requestBody := map[string]string{"activation_key": activationKey, "cluster_id": clusterID}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		go logger.Error(subApi.Acknowledgment, err)
		return nil, restErrors.NewInternalServerError("can't activate subscription")
	}

	bodyReader := bytes.NewReader(jsonBody)
	req, err := http.NewRequest(http.MethodPost, config.EnvironmentConf["SUBSCRIPTION_API_BASE_URL"]+ACKNOWLEDGEMENT, bodyReader)
	if err != nil {
		go logger.Error(subApi.Acknowledgment, err)
		return nil, restErrors.NewInternalServerError("can't activate subscription")
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{
		Timeout: 30 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		go logger.Error(subApi.Acknowledgment, err)
		return nil, restErrors.NewInternalServerError("can't activate subscription")
	}

	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusConflict {
			return nil, restErrors.NewConflictError("subscription already used by another cluster")
		}
		go logger.Error(subApi.Acknowledgment, errors.New(res.Status))
		return nil, restErrors.NewInternalServerError("can't activate subscription")
	}

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		go logger.Error(subApi.Acknowledgment, err)
		return nil, restErrors.NewInternalServerError("can't activate subscription")
	}

	return responseData, nil
}

func (subApi *subscriptionService) CurrentTimeStamp() ([]byte, *restErrors.RestErr) {

	req, err := http.NewRequest(http.MethodGet, config.EnvironmentConf["SUBSCRIPTION_API_BASE_URL"]+CURRENT_TIMESTAMP, nil)
	if err != nil {
		go logger.Error(subApi.CurrentTimeStamp, err)
		return nil, restErrors.NewInternalServerError(err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{
		Timeout: 30 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		go logger.Error(subApi.CurrentTimeStamp, err)
		return nil, restErrors.NewInternalServerError(err.Error())
	}

	if res.StatusCode != http.StatusOK {
		go logger.Error(subApi.CurrentTimeStamp, errors.New(res.Status))
		return nil, restErrors.NewInternalServerError(err.Error())
	}

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		go logger.Error(subApi.CurrentTimeStamp, err)
		return nil, restErrors.NewInternalServerError(err.Error())
	}

	return responseData, nil
}
