package subscription

import (
	"bytes"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/api/pkg/shared"
	subscriptionAPI "github.com/kotalco/cloud-api/pkg/subscription"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http/httptest"

	"net/http"
	"os"
	"testing"
)

/*
 subscriptionAPI service  mocks
*/
var (
	subscriptionAcknowledgmentFunc   func(activationKey string) *restErrors.RestErr
	subscriptionCurrentTimestampFunc func() (int64, *restErrors.RestErr)
)

type subscriptionServiceMock struct{}

func (s subscriptionServiceMock) Acknowledgment(activationKey string) *restErrors.RestErr {
	return subscriptionAcknowledgmentFunc(activationKey)
}
func (s subscriptionServiceMock) CurrentTimestamp() (int64, *restErrors.RestErr) {
	return subscriptionCurrentTimestampFunc()
}

func newFiberCtx(dto interface{}, method func(c *fiber.Ctx) error, locals map[string]interface{}) ([]byte, *http.Response) {
	app := fiber.New()
	app.Post("/test/", func(c *fiber.Ctx) error {
		for key, element := range locals {
			c.Locals(key, element)
		}
		return method(c)
	})

	marshaledDto, err := json.Marshal(dto)
	if err != nil {
		panic(err.Error())
	}

	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(marshaledDto))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	return body, resp
}

func TestMain(m *testing.M) {

	subscriptionService = &subscriptionServiceMock{}

	code := m.Run()
	os.Exit(code)
}

func TestAcknowledgement(t *testing.T) {
	var validDto = map[string]string{
		"activation_key": "1234",
	}
	var invalidDto = map[string]string{}
	t.Run("Acknowledgement should pass", func(t *testing.T) {
		subscriptionAcknowledgmentFunc = func(activationKey string) *restErrors.RestErr {
			return nil
		}

		subscriptionAPI.IsValid = func() bool {
			return true
		}
		body, resp := newFiberCtx(validDto, Acknowledgement, map[string]interface{}{})

		var result map[string]shared.SuccessMessage
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusOK, resp.StatusCode)

	})
	t.Run("Acknowledgement should throw a bad request error", func(t *testing.T) {
		body, resp := newFiberCtx("", Acknowledgement, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)

	})
	t.Run("Acknowledgement should throw a validation error", func(t *testing.T) {
		body, resp := newFiberCtx(invalidDto, Acknowledgement, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		if err != nil {
			panic(err.Error())
		}
		var fields = map[string]string{}
		fields["activation_key"] = "invalid key"
		badReqErr := restErrors.NewValidationError(fields)

		assert.EqualValues(t, badReqErr.Status, resp.StatusCode)
		assert.Equal(t, *badReqErr, result)

	})
	t.Run("Acknowledgement should throw if subscription acknowledgment throws", func(t *testing.T) {
		subscriptionAcknowledgmentFunc = func(activationKey string) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		body, resp := newFiberCtx(validDto, Acknowledgement, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)

	})
	t.Run("Acknowledgement should thrwo if subscription invalid", func(t *testing.T) {
		subscriptionAcknowledgmentFunc = func(activationKey string) *restErrors.RestErr {
			return nil
		}

		subscriptionAPI.IsValid = func() bool {
			return false
		}
		body, resp := newFiberCtx(validDto, Acknowledgement, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusGone, resp.StatusCode)

	})

}

func TestCurrent(t *testing.T) {
	t.Run("Current should pass", func(t *testing.T) {
		subscriptionAPI.SubscriptionDetails = &subscriptionAPI.SubscriptionDetailsDto{}

		body, resp := newFiberCtx("", Current, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
	})
	t.Run("Current should throw if subscription details doesn't exits", func(t *testing.T) {
		subscriptionAPI.SubscriptionDetails = nil

		body, resp := newFiberCtx("", Current, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusGone, resp.StatusCode)
	})

}
