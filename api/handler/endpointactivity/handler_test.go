package endpointactivity

import (
	"bytes"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/core/endpointactivity"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var (
	GetByEndpointIdFunc func(endpointId string) (*endpointactivity.Activity, restErrors.IRestErr)
	IncrementFunc       func(activity *endpointactivity.Activity) restErrors.IRestErr
)

type activityServiceMock struct{}

func (s activityServiceMock) WithoutTransaction() endpointactivity.IService {
	return s
}

func (s activityServiceMock) WithTransaction(txHandle *gorm.DB) endpointactivity.IService {
	return s
}
func (s activityServiceMock) GetByEndpointId(endpointId string) (*endpointactivity.Activity, restErrors.IRestErr) {
	return GetByEndpointIdFunc(endpointId)
}

func (s activityServiceMock) Increment(activity *endpointactivity.Activity) restErrors.IRestErr {
	return IncrementFunc(activity)
}

func TestMain(m *testing.M) {
	activityService = &activityServiceMock{}
	sqlclient.OpenDBConnection()
	code := m.Run()
	os.Exit(code)
}

func newFiberCtx(dto interface{}, method func(c *fiber.Ctx) error, locals map[string]interface{}) ([]byte, *http.Response) {
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
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

func TestLogs(t *testing.T) {
	var validDto = map[string]string{
		"request_id": "12345678",
	}
	var invalidDto = map[string]string{}

	t.Run("Logs_should_pass", func(t *testing.T) {
		GetByEndpointIdFunc = func(endpointId string) (*endpointactivity.Activity, restErrors.IRestErr) {
			return &endpointactivity.Activity{}, nil
		}
		IncrementFunc = func(activity *endpointactivity.Activity) restErrors.IRestErr {
			return nil
		}
		_, resp := newFiberCtx(validDto, Logs, map[string]interface{}{})

		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
	})
	t.Run("Logs_should_pass_and_create_new_endpoint_activity_if_it_doesn't_exist", func(t *testing.T) {
		GetByEndpointIdFunc = func(endpointId string) (*endpointactivity.Activity, restErrors.IRestErr) {
			return nil, restErrors.NewNotFoundError("no such record")
		}
		IncrementFunc = func(activity *endpointactivity.Activity) restErrors.IRestErr {
			return nil
		}
		_, resp := newFiberCtx(validDto, Logs, map[string]interface{}{})

		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
	})
	t.Run("Logs_should_throw_bad_request_err", func(t *testing.T) {
		GetByEndpointIdFunc = func(endpointId string) (*endpointactivity.Activity, restErrors.IRestErr) {
			return &endpointactivity.Activity{}, nil
		}
		IncrementFunc = func(activity *endpointactivity.Activity) restErrors.IRestErr {
			return nil
		}
		_, resp := newFiberCtx("", Logs, map[string]interface{}{})

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Logs_should_throw_validation_error", func(t *testing.T) {
		_, resp := newFiberCtx(invalidDto, Logs, map[string]interface{}{})
		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)

	})

	t.Run("Logs_should_throw_if_can't_increment_endpoint_counter", func(t *testing.T) {
		GetByEndpointIdFunc = func(endpointId string) (*endpointactivity.Activity, restErrors.IRestErr) {
			return &endpointactivity.Activity{}, nil
		}
		IncrementFunc = func(activity *endpointactivity.Activity) restErrors.IRestErr {
			return restErrors.NewInternalServerError("something went worng")
		}

		_, resp := newFiberCtx(validDto, Logs, map[string]interface{}{})
		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)
	})

}
