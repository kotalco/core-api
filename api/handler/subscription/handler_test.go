package subscription

import (
	"bytes"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/internal/setting"
	subscriptionAPI "github.com/kotalco/cloud-api/pkg/subscription"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/shared"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
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

/*
setting service  mocks
*/
var (
	settingSettingsFunc               func() ([]*setting.Setting, *restErrors.RestErr)
	settingConfigureDomainFunc        func(dto *setting.ConfigureDomainRequestDto) *restErrors.RestErr
	settingIsDomainConfiguredFunc     func() bool
	settingConfigureRegistrationFunc  func(dto *setting.ConfigureRegistrationRequestDto) *restErrors.RestErr
	settingIsRegistrationEnabledFunc  func() bool
	settingConfigureActivationKeyFunc func(key string) *restErrors.RestErr
	settingGetActivationKey           func() (string, *restErrors.RestErr)
)

type settingServiceMocks struct{}

func (s settingServiceMocks) ConfigureActivationKey(key string) *restErrors.RestErr {
	return settingConfigureActivationKeyFunc(key)
}

func (s settingServiceMocks) GetActivationKey() (string, *restErrors.RestErr) {
	return settingGetActivationKey()
}

func (s settingServiceMocks) ConfigureRegistration(dto *setting.ConfigureRegistrationRequestDto) *restErrors.RestErr {
	return settingConfigureRegistrationFunc(dto)
}

func (s settingServiceMocks) IsRegistrationEnabled() bool {
	return settingIsRegistrationEnabledFunc()
}

func (s settingServiceMocks) WithoutTransaction() setting.IService {
	return s
}

func (s settingServiceMocks) WithTransaction(txHandle *gorm.DB) setting.IService {
	return s
}

func (s settingServiceMocks) Settings() ([]*setting.Setting, *restErrors.RestErr) {
	return settingSettingsFunc()
}

func (s settingServiceMocks) ConfigureDomain(dto *setting.ConfigureDomainRequestDto) *restErrors.RestErr {
	return settingConfigureDomainFunc(dto)
}

func (s settingServiceMocks) IsDomainConfigured() bool {
	return settingIsDomainConfiguredFunc()
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
	settingService = &settingServiceMocks{}

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

		settingConfigureActivationKeyFunc = func(key string) *restErrors.RestErr {
			return nil
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

		assert.EqualValues(t, http.StatusForbidden, resp.StatusCode)

	})
	t.Run("Acknowledgement should throw if can't configure the activation key", func(t *testing.T) {
		subscriptionAcknowledgmentFunc = func(activationKey string) *restErrors.RestErr {
			return nil
		}

		subscriptionAPI.IsValid = func() bool {
			return true
		}

		settingConfigureActivationKeyFunc = func(key string) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}
		body, resp := newFiberCtx(validDto, Acknowledgement, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)
		assert.EqualValues(t, "can't save activation key", result.Message)

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

		assert.EqualValues(t, http.StatusForbidden, resp.StatusCode)
	})

}
