package setting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/internal/setting"
	"github.com/kotalco/cloud-api/pkg/token"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/shared"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

/*
setting service  mocks
*/
var (
	settingSettingsFunc              func() ([]*setting.Setting, *restErrors.RestErr)
	settingConfigureDomainFunc       func(dto *setting.ConfigureDomainRequestDto) *restErrors.RestErr
	settingIsDomainConfiguredFunc    func() bool
	settingConfigureRegistrationFunc func(dto *setting.ConfigureRegistrationRequestDto) *restErrors.RestErr
	settingIsRegistrationEnabledFunc func() bool
)

type settingServiceMocks struct{}

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
	settingService = &settingServiceMocks{}

	code := m.Run()
	os.Exit(code)
}
func TestConfigureDomain(t *testing.T) {
	userDetails := new(token.UserDetails)
	userDetails.ID = "test@test.com"
	var locals = map[string]interface{}{}
	locals["user"] = *userDetails
	var validDto = map[string]string{
		"domain": "kotal.co",
	}
	var invalidDto = map[string]string{}

	t.Run("ConfigureDomain should pass", func(t *testing.T) {
		settingConfigureDomainFunc = func(dto *setting.ConfigureDomainRequestDto) *restErrors.RestErr {
			return nil
		}
		body, resp := newFiberCtx(validDto, ConfigureDomain, locals)

		var result map[string]shared.SuccessMessage
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}
		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
		assert.EqualValues(t, "domain configured successfully!", result["data"].Message)
	})
	t.Run("configure domain should throw bad request err", func(t *testing.T) {
		body, resp := newFiberCtx("", ConfigureDomain, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)
		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "invalid request body", result.Message)
	})

	t.Run("configure domain should throw validation err", func(t *testing.T) {
		body, resp := newFiberCtx(invalidDto, ConfigureDomain, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)

		var fields = map[string]string{}
		fields["domain"] = "invalid domain"
		badReqErr := restErrors.NewValidationError(fields)

		assert.Equal(t, *badReqErr, result)
		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("configure domain should throw if service throws", func(t *testing.T) {
		settingConfigureDomainFunc = func(dto *setting.ConfigureDomainRequestDto) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}
		body, resp := newFiberCtx(validDto, ConfigureDomain, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)
		assert.EqualValues(t, http.StatusInternalServerError, result.Status)
		assert.EqualValues(t, "something went wrong", result.Message)
		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
func TestConfigureRegistration(t *testing.T) {
	userDetails := new(token.UserDetails)
	userDetails.ID = "test@test.com"
	var locals = map[string]interface{}{}
	locals["user"] = *userDetails
	var validDto = map[string]bool{
		"enable_registration": true,
	}
	var invalidDto = map[string]bool{}

	t.Run("Configure registration should pass", func(t *testing.T) {
		settingConfigureRegistrationFunc = func(dto *setting.ConfigureRegistrationRequestDto) *restErrors.RestErr {
			return nil
		}
		body, resp := newFiberCtx(validDto, ConfigureRegistration, locals)

		fmt.Println(string(body), resp)

		var result map[string]shared.SuccessMessage
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}
		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
		assert.EqualValues(t, "registration configured successfully!", result["data"].Message)
	})
	t.Run("configure registration should throw bad request err", func(t *testing.T) {
		body, resp := newFiberCtx("", ConfigureRegistration, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)
		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "invalid request body", result.Message)
	})

	t.Run("configure registration should throw validation err", func(t *testing.T) {
		body, resp := newFiberCtx(invalidDto, ConfigureRegistration, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)

		var fields = map[string]string{}
		fields["enable_registration"] = "invalid registration value"
		badReqErr := restErrors.NewValidationError(fields)

		assert.Equal(t, *badReqErr, result)
		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("configure registration should throw if service throws", func(t *testing.T) {
		settingConfigureRegistrationFunc = func(dto *setting.ConfigureRegistrationRequestDto) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}
		body, resp := newFiberCtx(validDto, ConfigureRegistration, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)
		assert.EqualValues(t, http.StatusInternalServerError, result.Status)
		assert.EqualValues(t, "something went wrong", result.Message)
		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
func TestSettings(t *testing.T) {
	userDetails := new(token.UserDetails)
	userDetails.ID = "test@test.com"
	var locals = map[string]interface{}{}
	locals["user"] = *userDetails
	t.Run("settings should pass", func(t *testing.T) {
		settingSettingsFunc = func() ([]*setting.Setting, *restErrors.RestErr) {
			return []*setting.Setting{{}}, nil
		}
		body, resp := newFiberCtx("", Settings, locals)

		var result map[string][]setting.SettingResponseDto
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}
		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
	})
	t.Run("setting should throw if service throws", func(t *testing.T) {
		settingSettingsFunc = func() ([]*setting.Setting, *restErrors.RestErr) {
			return nil, restErrors.NewInternalServerError("something went wrong")
		}
		body, resp := newFiberCtx("", Settings, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}
		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)
		assert.EqualValues(t, http.StatusInternalServerError, result.Status)
		assert.EqualValues(t, "something went wrong", result.Message)

	})

}
