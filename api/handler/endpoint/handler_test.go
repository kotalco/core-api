package endpoint

import (
	"bytes"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/internal/endpoint"
	"github.com/kotalco/cloud-api/internal/setting"
	"github.com/kotalco/cloud-api/internal/workspace"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/shared"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

/*
endpoint service mocks
*/
var (
	endpointServiceCreateFunc func(dto *endpoint.CreateEndpointDto, svc *corev1.Service) *restErrors.RestErr
	endpointServiceListFunc   func(namespace string) ([]*endpoint.EndpointDto, *restErrors.RestErr)
	endpointServiceGetFunc    func(name string, namespace string) (*endpoint.EndpointDto, *restErrors.RestErr)
	endpointServiceDeleteFunc func(name string, namespace string) *restErrors.RestErr
)

type endpointServiceMock struct{}

func (e endpointServiceMock) Create(dto *endpoint.CreateEndpointDto, svc *corev1.Service) *restErrors.RestErr {
	return endpointServiceCreateFunc(dto, svc)
}
func (e endpointServiceMock) List(namespace string) ([]*endpoint.EndpointDto, *restErrors.RestErr) {
	return endpointServiceListFunc(namespace)
}
func (e endpointServiceMock) Get(name string, namespace string) (*endpoint.EndpointDto, *restErrors.RestErr) {
	return endpointServiceGetFunc(name, namespace)
}
func (e endpointServiceMock) Delete(name string, namespace string) *restErrors.RestErr {
	return endpointServiceDeleteFunc(name, namespace)
}

/*
svc service mock
*/

var (
	svcServiceListFunc func(namespace string) (*corev1.ServiceList, *restErrors.RestErr)
	svcServiceGetFunc  func(name string, namespace string) (*corev1.Service, *restErrors.RestErr)
)

type svcServiceMock struct{}

func (s svcServiceMock) List(namespace string) (*corev1.ServiceList, *restErrors.RestErr) {
	return svcServiceListFunc(namespace)
}

func (s svcServiceMock) Get(name string, namespace string) (*corev1.Service, *restErrors.RestErr) {
	return svcServiceGetFunc(name, namespace)
}

var (
	settingWithTransaction               func(txHandle *gorm.DB) setting.IService
	settingSettingsFunc                  func() (*setting.Setting, *restErrors.RestErr)
	settingConfigureDomainFunc           func(dto *setting.ConfigureDomainRequestDto) *restErrors.RestErr
	settingUpdateDomainConfigurationFunc func(dto *setting.ConfigureDomainRequestDto) *restErrors.RestErr
	settingIsDomainConfiguredFunc        func() bool
)

type settingServiceMock struct{}

func (s settingServiceMock) WithTransaction(txHandle *gorm.DB) setting.IService {
	return s
}

func (s settingServiceMock) Settings() (*setting.Setting, *restErrors.RestErr) {
	return settingSettingsFunc()
}

func (s settingServiceMock) ConfigureDomain(dto *setting.ConfigureDomainRequestDto) *restErrors.RestErr {
	return settingConfigureDomainFunc(dto)
}

func (s settingServiceMock) UpdateDomainConfiguration(dto *setting.ConfigureDomainRequestDto) *restErrors.RestErr {
	return settingUpdateDomainConfigurationFunc(dto)
}

func (s settingServiceMock) IsDomainConfigured() bool {
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
	endpointService = &endpointServiceMock{}
	svcService = &svcServiceMock{}
	settingService = &settingServiceMock{}
	code := m.Run()

	os.Exit(code)
}

func TestCreate(t *testing.T) {
	workspaceModel := new(workspace.Workspace)
	var locals = map[string]interface{}{}
	locals["workspace"] = *workspaceModel

	var validDto = map[string]string{
		"name":         "name",
		"service_name": "serviceName",
	}

	var invalidDto = map[string]string{
		"name": "name",
	}

	t.Run("create endpoint should pass", func(t *testing.T) {
		settingIsDomainConfiguredFunc = func() bool {
			return true
		}
		svcServiceGetFunc = func(name string, namespace string) (*corev1.Service, *restErrors.RestErr) {
			return &corev1.Service{Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{{}}}}, nil
		}
		endpointServiceCreateFunc = func(dto *endpoint.CreateEndpointDto, svc *corev1.Service) *restErrors.RestErr {
			return nil
		}
		availableProtocol = func(protocol string) bool {
			return true
		}
		body, resp := newFiberCtx(validDto, Create, locals)
		var result map[string]shared.SuccessMessage
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)

		assert.EqualValues(t, http.StatusCreated, resp.StatusCode)
		assert.EqualValues(t, "Endpoint has been created", result["data"].Message)
	})

	t.Run("create endpoint should throw bad request err", func(t *testing.T) {
		body, resp := newFiberCtx("", Create, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)
		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "invalid request body", result.Message)
	})

	t.Run("create should throw validation err", func(t *testing.T) {
		body, resp := newFiberCtx(invalidDto, Create, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)

		var fields = map[string]string{}
		fields["service_name"] = "invalid service_name"
		badReqErr := restErrors.NewValidationError(fields)

		assert.Equal(t, *badReqErr, result)
		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
	})
	t.Run("create endpoint should throw if the user didn't configure the domain based url", func(t *testing.T) {
		settingIsDomainConfiguredFunc = func() bool {
			return false
		}

		body, resp := newFiberCtx(validDto, Create, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)

		assert.EqualValues(t, http.StatusForbidden, resp.StatusCode)
		assert.EqualValues(t, "Domain hasn't been configured yet !", result.Message)
	})

	t.Run("create should throw if svcService.Get can't find service", func(t *testing.T) {
		settingIsDomainConfiguredFunc = func() bool {
			return true
		}
		svcServiceGetFunc = func(name string, namespace string) (*corev1.Service, *restErrors.RestErr) {
			return nil, restErrors.NewNotFoundError("no such record")
		}

		body, resp := newFiberCtx(validDto, Create, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)

		assert.EqualValues(t, http.StatusNotFound, resp.StatusCode)
		assert.EqualValues(t, "no such record", result.Message)
	})

	t.Run("create endpoint should throw if can't endpointService.create throws", func(t *testing.T) {
		settingIsDomainConfiguredFunc = func() bool {
			return true
		}
		svcServiceGetFunc = func(name string, namespace string) (*corev1.Service, *restErrors.RestErr) {
			return &corev1.Service{Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{{}}}}, nil
		}
		endpointServiceCreateFunc = func(dto *endpoint.CreateEndpointDto, svc *corev1.Service) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}
		availableProtocol = func(protocol string) bool {
			return true
		}
		body, resp := newFiberCtx(validDto, Create, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)

		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)
		assert.EqualValues(t, "something went wrong", result.Message)
	})
	t.Run("create endpoint should throw if there is no valid protocols", func(t *testing.T) {
		settingIsDomainConfiguredFunc = func() bool {
			return true
		}
		svcServiceGetFunc = func(name string, namespace string) (*corev1.Service, *restErrors.RestErr) {
			return &corev1.Service{Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{{}}}}, nil
		}
		availableProtocol = func(protocol string) bool {
			return false
		}
		body, resp := newFiberCtx(validDto, Create, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "service  doesn't have API enabled", result.Message)
	})

}

func TestList(t *testing.T) {
	workspaceModel := new(workspace.Workspace)
	var locals = map[string]interface{}{}
	locals["workspace"] = *workspaceModel

	t.Run("list endpoints should pass", func(t *testing.T) {
		endpointServiceListFunc = func(namespace string) ([]*endpoint.EndpointDto, *restErrors.RestErr) {
			return []*endpoint.EndpointDto{{}}, nil
		}

		body, resp := newFiberCtx("", List, locals)
		var result map[string][]*endpoint.EndpointDto
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)

		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
		assert.NotNil(t, result["data"])
	})

	t.Run("list endpoint should throw if endpointServiceList throws", func(t *testing.T) {
		endpointServiceListFunc = func(namespace string) ([]*endpoint.EndpointDto, *restErrors.RestErr) {
			return nil, restErrors.NewInternalServerError("something went wrong")
		}

		body, resp := newFiberCtx("", List, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)
		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)
		assert.EqualValues(t, "something went wrong", result.Message)
	})

}

func TestGet(t *testing.T) {
	workspaceModel := new(workspace.Workspace)
	var locals = map[string]interface{}{}
	locals["workspace"] = *workspaceModel

	t.Run("get endpoint should pass", func(t *testing.T) {
		endpointServiceGetFunc = func(name string, namespace string) (*endpoint.EndpointDto, *restErrors.RestErr) {
			return &endpoint.EndpointDto{}, nil
		}
		body, resp := newFiberCtx("", Get, locals)
		var result map[string]endpoint.EndpointDto
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)
		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
		assert.NotNil(t, result["data"])

	})

	t.Run("get endpoint should throw if can't ger endpoint from service", func(t *testing.T) {
		endpointServiceGetFunc = func(name string, namespace string) (*endpoint.EndpointDto, *restErrors.RestErr) {
			return nil, restErrors.NewNotFoundError("no such record")
		}
		body, resp := newFiberCtx("", Get, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)
		assert.EqualValues(t, http.StatusNotFound, resp.StatusCode)
		assert.NotNil(t, "no such record", result.Message)
	})

}

func TestDelete(t *testing.T) {
	workspaceModel := new(workspace.Workspace)
	var locals = map[string]interface{}{}
	locals["workspace"] = *workspaceModel

	t.Run("delete endpoint should pass", func(t *testing.T) {
		endpointServiceDeleteFunc = func(name string, namespace string) *restErrors.RestErr {
			return nil
		}
		body, resp := newFiberCtx("", Delete, locals)
		var result map[string]shared.SuccessMessage
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)
		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
		assert.EqualValues(t, "Endpoint Deleted", result["data"].Message)

	})

	t.Run("delete endpoint should throw if can't delete endpoint from service", func(t *testing.T) {
		endpointServiceDeleteFunc = func(name string, namespace string) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}
		body, resp := newFiberCtx("", Delete, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)
		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)
		assert.NotNil(t, "something went wrong", result.Message)
	})

}
