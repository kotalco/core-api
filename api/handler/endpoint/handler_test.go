package endpoint

import (
	"bytes"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/core-api/core/endpoint"
	"github.com/kotalco/core-api/core/endpointactivity"
	"github.com/kotalco/core-api/core/setting"
	"github.com/kotalco/core-api/core/workspace"
	"github.com/kotalco/core-api/k8s/secret"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/responder"
	"github.com/kotalco/core-api/pkg/token"
	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	"gorm.io/gorm"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

/*
endpoint service mocks
*/
var (
	endpointServiceCreateFunc func(dto *endpoint.CreateEndpointDto, svc *corev1.Service) restErrors.IRestErr
	endpointServiceListFunc   func(ns string, labels map[string]string) (*v1alpha1.IngressRouteList, restErrors.IRestErr)
	endpointServiceGetFunc    func(name string, namespace string) (*v1alpha1.IngressRoute, restErrors.IRestErr)
	endpointServiceDeleteFunc func(name string, namespace string) restErrors.IRestErr
	endpointServiceCountFunc  func(ns string, labels map[string]string) (int, restErrors.IRestErr)
)

type endpointServiceMock struct{}

func (e endpointServiceMock) Create(dto *endpoint.CreateEndpointDto, svc *corev1.Service) restErrors.IRestErr {
	return endpointServiceCreateFunc(dto, svc)
}
func (e endpointServiceMock) List(ns string, labels map[string]string) (*v1alpha1.IngressRouteList, restErrors.IRestErr) {
	return endpointServiceListFunc(ns, labels)
}
func (e endpointServiceMock) Get(name string, namespace string) (*v1alpha1.IngressRoute, restErrors.IRestErr) {
	return endpointServiceGetFunc(name, namespace)
}
func (e endpointServiceMock) Delete(name string, namespace string) restErrors.IRestErr {
	return endpointServiceDeleteFunc(name, namespace)
}
func (e endpointServiceMock) Count(ns string, labels map[string]string) (int, restErrors.IRestErr) {
	return endpointServiceCountFunc(ns, labels)
}

/*
svc service mock
*/

var (
	svcServiceListFunc   func(namespace string) (*corev1.ServiceList, restErrors.IRestErr)
	svcServiceGetFunc    func(name string, namespace string) (*corev1.Service, restErrors.IRestErr)
	svcServiceCreateFunc func(obj *corev1.Service) restErrors.IRestErr
)

type svcServiceMock struct{}

func (s svcServiceMock) List(namespace string) (*corev1.ServiceList, restErrors.IRestErr) {
	return svcServiceListFunc(namespace)
}

func (s svcServiceMock) Get(name string, namespace string) (*corev1.Service, restErrors.IRestErr) {
	return svcServiceGetFunc(name, namespace)
}

func (s svcServiceMock) Create(obj *corev1.Service) restErrors.IRestErr {
	return svcServiceCreateFunc(obj)
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

type secretServiceMock struct{}

var (
	secretCreateFunc func(dto *secret.CreateSecretDto) restErrors.IRestErr
	secretGetFunc    func(name string, namespace string) (*corev1.Secret, restErrors.IRestErr)
)

func (s secretServiceMock) Create(dto *secret.CreateSecretDto) restErrors.IRestErr {
	return secretCreateFunc(dto)
}
func (s secretServiceMock) Get(name string, namespace string) (*corev1.Secret, restErrors.IRestErr) {
	return secretGetFunc(name, namespace)
}

var (
	activityCreateFunc func([]endpointactivity.CreateEndpointActivityDto) restErrors.IRestErr
	activityStatsFunc  func(startDate time.Time, endDate time.Time, endpointId string) (*[]endpointactivity.ActivityAggregations, restErrors.IRestErr)
)

type activityServiceMock struct{}

func (s activityServiceMock) WithoutTransaction() endpointactivity.IService {
	return s
}
func (s activityServiceMock) WithTransaction(txHandle *gorm.DB) endpointactivity.IService {
	return s
}

func (s activityServiceMock) Stats(startDate time.Time, endDate time.Time, endpointId string) (*[]endpointactivity.ActivityAggregations, restErrors.IRestErr) {
	return activityStatsFunc(startDate, endDate, endpointId)
}
func (s activityServiceMock) Create(dto []endpointactivity.CreateEndpointActivityDto) restErrors.IRestErr {
	return activityCreateFunc(dto)
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
	secretService = &secretServiceMock{}
	activityService = &activityServiceMock{}
	code := m.Run()

	os.Exit(code)
}

func TestCreate(t *testing.T) {
	workspaceModel := new(workspace.Workspace)
	var locals = map[string]interface{}{}
	locals["workspace"] = *workspaceModel
	userDetails := new(token.UserDetails)
	userDetails.ID = "test@test.com"
	locals["user"] = *userDetails

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
		var result map[string]responder.SuccessMessage
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
		endpointServiceListFunc = func(ns string, labels map[string]string) (*v1alpha1.IngressRouteList, restErrors.IRestErr) {
			return &v1alpha1.IngressRouteList{Items: []v1alpha1.IngressRoute{{}}}, nil
		}

		body, resp := newFiberCtx("", List, locals)
		var result map[string][]*endpoint.EndpointMetaDto
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)

		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
		assert.NotNil(t, result["data"])
	})

	t.Run("list endpoint should throw if endpointServiceList throws", func(t *testing.T) {
		endpointServiceListFunc = func(ns string, labels map[string]string) (*v1alpha1.IngressRouteList, restErrors.IRestErr) {
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
		endpointServiceGetFunc = func(name string, namespace string) (*v1alpha1.IngressRoute, restErrors.IRestErr) {
			return &v1alpha1.IngressRoute{}, nil
		}
		secretGetFunc = func(name string, namespace string) (*corev1.Secret, restErrors.IRestErr) {
			return &corev1.Secret{}, nil
		}
		body, resp := newFiberCtx("", Get, locals)
		var result map[string]endpoint.EndpointDto
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)
		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
		assert.NotNil(t, result["data"])

	})

	t.Run("get endpoint should throw if can't ger endpoint from service", func(t *testing.T) {
		endpointServiceGetFunc = func(name string, namespace string) (*v1alpha1.IngressRoute, restErrors.IRestErr) {
			return nil, restErrors.NewNotFoundError("no such record")
		}
		body, resp := newFiberCtx("", Get, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)
		assert.EqualValues(t, http.StatusNotFound, resp.StatusCode)
		assert.NotNil(t, "no such record", result.Message)
	})

	t.Run("get endpoint shouldnot throw if get secret fails coz it means this endpoint has no secret", func(t *testing.T) {
		endpointServiceGetFunc = func(name string, namespace string) (*v1alpha1.IngressRoute, restErrors.IRestErr) {
			return &v1alpha1.IngressRoute{}, nil
		}
		secretGetFunc = func(name string, namespace string) (*corev1.Secret, restErrors.IRestErr) {
			return nil, restErrors.NewNotFoundError("no such record")
		}
		body, resp := newFiberCtx("", Get, locals)
		var result map[string]endpoint.EndpointDto
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)
		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
		assert.NotNil(t, result["data"])

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
		var result map[string]responder.SuccessMessage
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

func TestCount(t *testing.T) {
	workspaceModel := new(workspace.Workspace)
	var locals = map[string]interface{}{}
	locals["workspace"] = *workspaceModel

	t.Run("count endpoints should pass", func(t *testing.T) {
		endpointServiceCountFunc = func(ns string, labels map[string]string) (int, restErrors.IRestErr) {
			return 1, nil
		}
		_, resp := newFiberCtx("", Count, locals)
		assert.EqualValues(t, resp.Header.Get("X-Total-Count"), "1")
		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("count endpoint should throw if service throws", func(t *testing.T) {
		endpointServiceCountFunc = func(ns string, labels map[string]string) (int, restErrors.IRestErr) {
			return 0, restErrors.NewInternalServerError("something went wrong")
		}
		_, resp := newFiberCtx("", Count, locals)
		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestWriteStats(t *testing.T) {
	var validDto = []map[string]interface{}{{"request_id": "12345678", "count": 1}}
	var invalidDto = []map[string]interface{}{{}}

	t.Run("WriteStats_should_pass", func(t *testing.T) {
		activityCreateFunc = func([]endpointactivity.CreateEndpointActivityDto) restErrors.IRestErr {
			return nil
		}
		_, resp := newFiberCtx(validDto, WriteStats, map[string]interface{}{})

		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
	})
	t.Run("WriteStats_should_throw_bad_request_err", func(t *testing.T) {
		activityCreateFunc = func([]endpointactivity.CreateEndpointActivityDto) restErrors.IRestErr {
			return nil
		}
		_, resp := newFiberCtx("", WriteStats, map[string]interface{}{})

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("WriteStats_should_throw_validation_error", func(t *testing.T) {
		_, resp := newFiberCtx(invalidDto, WriteStats, map[string]interface{}{})
		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)

	})

	t.Run("WriteStats_should_throw_if_can't_create_endpoint_activity", func(t *testing.T) {
		activityCreateFunc = func([]endpointactivity.CreateEndpointActivityDto) restErrors.IRestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		_, resp := newFiberCtx(validDto, WriteStats, map[string]interface{}{})
		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)
	})

}
