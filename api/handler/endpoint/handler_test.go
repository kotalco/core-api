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
	endpointServiceCreateFunc func(dto *endpoint.CreateEndpointDto, svc *corev1.Service) restErrors.IRestErr
	endpointServiceListFunc   func(namespace string) ([]*endpoint.EndpointMetaDto, restErrors.IRestErr)
	endpointServiceGetFunc    func(name string, namespace string) (*endpoint.EndpointDto, restErrors.IRestErr)
	endpointServiceDeleteFunc func(name string, namespace string) restErrors.IRestErr
	endpointServiceCountFunc  func(namespace string) (int, restErrors.IRestErr)
)

type endpointServiceMock struct{}

func (e endpointServiceMock) Create(dto *endpoint.CreateEndpointDto, svc *corev1.Service) restErrors.IRestErr {
	return endpointServiceCreateFunc(dto, svc)
}
func (e endpointServiceMock) List(namespace string) ([]*endpoint.EndpointMetaDto, restErrors.IRestErr) {
	return endpointServiceListFunc(namespace)
}
func (e endpointServiceMock) Get(name string, namespace string) (*endpoint.EndpointDto, restErrors.IRestErr) {
	return endpointServiceGetFunc(name, namespace)
}
func (e endpointServiceMock) Delete(name string, namespace string) restErrors.IRestErr {
	return endpointServiceDeleteFunc(name, namespace)
}
func (e endpointServiceMock) Count(namespace string) (int, restErrors.IRestErr) {
	return endpointServiceCountFunc(namespace)
}

/*
svc service mock
*/

var (
	svcServiceListFunc func(namespace string) (*corev1.ServiceList, restErrors.IRestErr)
	svcServiceGetFunc  func(name string, namespace string) (*corev1.Service, restErrors.IRestErr)
)

type svcServiceMock struct{}

func (s svcServiceMock) List(namespace string) (*corev1.ServiceList, restErrors.IRestErr) {
	return svcServiceListFunc(namespace)
}

func (s svcServiceMock) Get(name string, namespace string) (*corev1.Service, restErrors.IRestErr) {
	return svcServiceGetFunc(name, namespace)
}

var (
	settingWithTransaction            func(txHandle *gorm.DB) setting.IService
	settingSettingsFunc               func() ([]*setting.Setting, restErrors.IRestErr)
	settingConfigureDomainFunc        func(dto *setting.ConfigureDomainRequestDto) restErrors.IRestErr
	settingIsDomainConfiguredFunc     func() bool
	settingConfigureRegistrationFunc  func(dto *setting.ConfigureRegistrationRequestDto) restErrors.IRestErr
	settingIsRegistrationEnabledFunc  func() bool
	settingConfigureActivationKeyFunc func(key string) restErrors.IRestErr
	settingGetActivationKeyFunc       func() (string, restErrors.IRestErr)
)

type settingServiceMock struct{}

func (s settingServiceMock) ConfigureActivationKey(key string) restErrors.IRestErr {
	return settingConfigureActivationKeyFunc(key)
}

func (s settingServiceMock) GetActivationKey() (string, restErrors.IRestErr) {
	return settingGetActivationKeyFunc()
}

func (s settingServiceMock) ConfigureRegistration(dto *setting.ConfigureRegistrationRequestDto) restErrors.IRestErr {
	return settingConfigureRegistrationFunc(dto)
}

func (s settingServiceMock) IsRegistrationEnabled() bool {
	return settingIsRegistrationEnabledFunc()
}

func (s settingServiceMock) WithoutTransaction() setting.IService {
	return s
}

func (s settingServiceMock) WithTransaction(txHandle *gorm.DB) setting.IService {
	return s
}

func (s settingServiceMock) Settings() ([]*setting.Setting, restErrors.IRestErr) {
	return settingSettingsFunc()
}

func (s settingServiceMock) ConfigureDomain(dto *setting.ConfigureDomainRequestDto) restErrors.IRestErr {
	return settingConfigureDomainFunc(dto)
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
		svcServiceGetFunc = func(name string, namespace string) (*corev1.Service, restErrors.IRestErr) {
			return &corev1.Service{Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{{}}}}, nil
		}
		endpointServiceCreateFunc = func(dto *endpoint.CreateEndpointDto, svc *corev1.Service) restErrors.IRestErr {
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

		assert.Equal(t, badReqErr, result)
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
		svcServiceGetFunc = func(name string, namespace string) (*corev1.Service, restErrors.IRestErr) {
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
		svcServiceGetFunc = func(name string, namespace string) (*corev1.Service, restErrors.IRestErr) {
			return &corev1.Service{Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{{}}}}, nil
		}
		endpointServiceCreateFunc = func(dto *endpoint.CreateEndpointDto, svc *corev1.Service) restErrors.IRestErr {
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
		svcServiceGetFunc = func(name string, namespace string) (*corev1.Service, restErrors.IRestErr) {
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
		endpointServiceListFunc = func(namespace string) ([]*endpoint.EndpointMetaDto, restErrors.IRestErr) {
			return []*endpoint.EndpointMetaDto{{}}, nil
		}

		body, resp := newFiberCtx("", List, locals)
		var result map[string][]*endpoint.EndpointMetaDto
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)

		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
		assert.NotNil(t, result["data"])
	})

	t.Run("list endpoint should throw if endpointServiceList throws", func(t *testing.T) {
		endpointServiceListFunc = func(namespace string) ([]*endpoint.EndpointMetaDto, restErrors.IRestErr) {
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
		endpointServiceGetFunc = func(name string, namespace string) (*endpoint.EndpointDto, restErrors.IRestErr) {
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
		endpointServiceGetFunc = func(name string, namespace string) (*endpoint.EndpointDto, restErrors.IRestErr) {
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
		endpointServiceDeleteFunc = func(name string, namespace string) restErrors.IRestErr {
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
		endpointServiceDeleteFunc = func(name string, namespace string) restErrors.IRestErr {
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
