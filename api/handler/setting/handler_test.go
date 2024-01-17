package setting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/core-api/core/setting"
	"github.com/kotalco/core-api/k8s/ingressroute"
	"github.com/kotalco/core-api/k8s/middleware"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/pagination"
	"github.com/kotalco/core-api/pkg/token"
	"github.com/stretchr/testify/assert"
	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	"gorm.io/gorm"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

/*
setting service  mocks
*/
var (
	settingSettingsFunc               func() ([]*setting.Setting, restErrors.IRestErr)
	settingConfigureDomainFunc        func(dto *setting.ConfigureDomainRequestDto) restErrors.IRestErr
	settingIsDomainConfiguredFunc     func() bool
	settingConfigureRegistrationFunc  func(dto *setting.ConfigureRegistrationRequestDto) restErrors.IRestErr
	settingIsRegistrationEnabledFunc  func() bool
	settingConfigureActivationKeyFunc func(key string) restErrors.IRestErr
	settingGetActivationKey           func() (string, restErrors.IRestErr)
)

type settingServiceMocks struct{}

func (s settingServiceMocks) ConfigureActivationKey(key string) restErrors.IRestErr {
	return settingConfigureActivationKeyFunc(key)
}

func (s settingServiceMocks) GetActivationKey() (string, restErrors.IRestErr) {
	return settingGetActivationKey()
}

func (s settingServiceMocks) ConfigureRegistration(dto *setting.ConfigureRegistrationRequestDto) restErrors.IRestErr {
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

func (s settingServiceMocks) Settings() ([]*setting.Setting, restErrors.IRestErr) {
	return settingSettingsFunc()
}

func (s settingServiceMocks) ConfigureDomain(dto *setting.ConfigureDomainRequestDto) restErrors.IRestErr {
	return settingConfigureDomainFunc(dto)
}

func (s settingServiceMocks) IsDomainConfigured() bool {
	return settingIsDomainConfiguredFunc()
}

var (
	k8svcListFunc        func(namespace string) (*corev1.ServiceList, restErrors.IRestErr)
	k8svcGetFunc         func(name string, namespace string) (*corev1.Service, restErrors.IRestErr)
	svcServiceCreateFunc func(obj *corev1.Service) restErrors.IRestErr
)

type k8sServiceMock struct{}

func (k *k8sServiceMock) List(namespace string) (*corev1.ServiceList, restErrors.IRestErr) {
	return k8svcListFunc(namespace)
}

func (k *k8sServiceMock) Get(name string, namespace string) (*corev1.Service, restErrors.IRestErr) {
	return k8svcGetFunc(name, namespace)
}
func (k *k8sServiceMock) Create(obj *corev1.Service) restErrors.IRestErr {
	return svcServiceCreateFunc(obj)
}

var (
	ingressRouteCreateFunc func(dto *ingressroute.IngressRouteDto) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr)
	ingressRouteListFunc   func(ns string, labels map[string]string) (*traefikv1alpha1.IngressRouteList, restErrors.IRestErr)
	ingressRouteGetFunc    func(name string, namespace string) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr)
	ingressRouteDeleteFunc func(name string, namespace string) restErrors.IRestErr
	ingressRouteUpdateFunc func(record *traefikv1alpha1.IngressRoute) restErrors.IRestErr
)

type ingressRouteServiceMock struct{}

func (i ingressRouteServiceMock) Create(dto *ingressroute.IngressRouteDto) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr) {
	return ingressRouteCreateFunc(dto)
}

func (i ingressRouteServiceMock) List(ns string, labels map[string]string) (*traefikv1alpha1.IngressRouteList, restErrors.IRestErr) {
	return ingressRouteListFunc(ns, labels)
}

func (i ingressRouteServiceMock) Get(name string, namespace string) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr) {
	return ingressRouteGetFunc(name, namespace)
}

func (i ingressRouteServiceMock) Delete(name string, namespace string) restErrors.IRestErr {
	return ingressRouteDeleteFunc(name, namespace)
}
func (i ingressRouteServiceMock) Update(record *traefikv1alpha1.IngressRoute) restErrors.IRestErr {
	return ingressRouteUpdateFunc(record)
}

type k8MiddlewareServiceMock struct{}

var (
	k8middlewareCreateFunc func(dto *middleware.CreateMiddlewareDto) restErrors.IRestErr
)

func (k k8MiddlewareServiceMock) Create(dto *middleware.CreateMiddlewareDto) restErrors.IRestErr {
	return k8middlewareCreateFunc(dto)
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
	k8service = &k8sServiceMock{}
	ingressRouteService = &ingressRouteServiceMock{}

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

	tmpNetworkIdentifiers := networkIdentifiers
	tmpVerifyDomainIp := verifyDomainIP
	tmpVerifyDomainHostName := verifyDomainHostName

	var invalidDto = map[string]string{}

	t.Run("ConfigureDomain should pass with ip address", func(t *testing.T) {
		networkIdentifiers = func() (ip string, hostName string, restErr restErrors.IRestErr) {
			return "1223", "", nil
		}
		verifyDomainIP = func(domain string, ipAddress string) restErrors.IRestErr {
			return nil
		}
		settingConfigureDomainFunc = func(dto *setting.ConfigureDomainRequestDto) restErrors.IRestErr {
			return nil
		}

		ingressRouteGetFunc = func(name string, namespace string) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr) {
			return &traefikv1alpha1.IngressRoute{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec: traefikv1alpha1.IngressRouteSpec{
					Routes: []traefikv1alpha1.Route{{
						Services: []traefikv1alpha1.Service{{
							LoadBalancerSpec: traefikv1alpha1.LoadBalancerSpec{Name: "kotal-dashboard"},
						}, {
							LoadBalancerSpec: traefikv1alpha1.LoadBalancerSpec{Name: "kotal-api"},
						}},
					}},
				},
			}, nil
		}

		ingressRouteUpdateFunc = func(record *traefikv1alpha1.IngressRoute) restErrors.IRestErr {
			return nil
		}
		body, resp := newFiberCtx(validDto, ConfigureDomain, locals)

		var result map[string]pagination.SuccessMessage
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}
		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
		assert.EqualValues(t, "domain configured successfully!", result["data"].Message)
	})
	t.Run("ConfigureDomain should pass with hostname", func(t *testing.T) {
		networkIdentifiers = func() (ip string, hostName string, restErr restErrors.IRestErr) {
			return "", "hostname.amazon.com", nil
		}
		verifyDomainHostName = func(domain string, hostname string) restErrors.IRestErr {
			return nil
		}
		settingConfigureDomainFunc = func(dto *setting.ConfigureDomainRequestDto) restErrors.IRestErr {
			return nil
		}

		ingressRouteGetFunc = func(name string, namespace string) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr) {
			return &traefikv1alpha1.IngressRoute{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec: traefikv1alpha1.IngressRouteSpec{
					Routes: []traefikv1alpha1.Route{{
						Services: []traefikv1alpha1.Service{{
							LoadBalancerSpec: traefikv1alpha1.LoadBalancerSpec{Name: "kotal-dashboard"},
						}, {
							LoadBalancerSpec: traefikv1alpha1.LoadBalancerSpec{Name: "kotal-api"},
						}},
					}},
				},
			}, nil
		}

		ingressRouteUpdateFunc = func(record *traefikv1alpha1.IngressRoute) restErrors.IRestErr {
			return nil
		}
		body, resp := newFiberCtx(validDto, ConfigureDomain, locals)

		var result map[string]pagination.SuccessMessage
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

		assert.Equal(t, badReqErr, result)
		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
	})
	t.Run("ConfigureDomain should throw if can't get ip address or hostname", func(t *testing.T) {
		networkIdentifiers = func() (ip string, hostname string, restErr restErrors.IRestErr) {
			return "", "", restErrors.NewInternalServerError("can't get ip address")
		}

		body, resp := newFiberCtx(validDto, ConfigureDomain, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}
		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)
		assert.EqualValues(t, "can't get ip address", result.Message)
	})
	t.Run("ConfigureDomain should throw if can't verify domain ip", func(t *testing.T) {
		networkIdentifiers = func() (ip string, hostName string, restErr restErrors.IRestErr) {
			return "1.2.3", "", nil
		}

		verifyDomainIP = func(domain string, ipAddress string) restErrors.IRestErr {
			return restErrors.NewInternalServerError("can't verify domain")
		}
		body, resp := newFiberCtx(validDto, ConfigureDomain, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}
		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)
		assert.EqualValues(t, "can't verify domain", result.Message)
	})
	t.Run("ConfigureDomain should throw if can't verify domain hostname", func(t *testing.T) {
		networkIdentifiers = func() (ip string, hostName string, restErr restErrors.IRestErr) {
			return "", "111.amazon.com", nil
		}

		verifyDomainHostName = func(domain string, hostName string) restErrors.IRestErr {
			return restErrors.NewInternalServerError("can't verify domain")
		}
		body, resp := newFiberCtx(validDto, ConfigureDomain, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}
		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)
		assert.EqualValues(t, "can't verify domain", result.Message)
	})
	t.Run("configure domain should throw if service throws", func(t *testing.T) {
		networkIdentifiers = func() (ip string, hostName string, restErr restErrors.IRestErr) {
			return "1223", "", nil
		}
		verifyDomainIP = func(domain string, ipAddress string) restErrors.IRestErr {
			return nil
		}
		settingConfigureDomainFunc = func(dto *setting.ConfigureDomainRequestDto) restErrors.IRestErr {
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
	t.Run("ConfigureDomain should throw if can't get kotal stack ingress-route", func(t *testing.T) {
		networkIdentifiers = func() (ip string, hostname string, restErr restErrors.IRestErr) {
			return "1223", "", nil
		}
		verifyDomainIP = func(domain string, ipAddress string) restErrors.IRestErr {
			return nil
		}
		settingConfigureDomainFunc = func(dto *setting.ConfigureDomainRequestDto) restErrors.IRestErr {
			return nil
		}

		ingressRouteGetFunc = func(name string, namespace string) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr) {
			return nil, restErrors.NewNotFoundError("no such record")
		}

		body, resp := newFiberCtx(validDto, ConfigureDomain, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}
		assert.EqualValues(t, http.StatusNotFound, resp.StatusCode)
		assert.EqualValues(t, "no such record", result.Message)
	})
	t.Run("ConfigureDomain should throw if can't update kotal stack", func(t *testing.T) {
		networkIdentifiers = func() (ip string, hostName string, restErr restErrors.IRestErr) {
			return "1223", "", nil
		}
		verifyDomainIP = func(domain string, ipAddress string) restErrors.IRestErr {
			return nil
		}
		settingConfigureDomainFunc = func(dto *setting.ConfigureDomainRequestDto) restErrors.IRestErr {
			return nil
		}

		ingressRouteGetFunc = func(name string, namespace string) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr) {
			return &traefikv1alpha1.IngressRoute{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec: traefikv1alpha1.IngressRouteSpec{
					Routes: []traefikv1alpha1.Route{{
						Services: []traefikv1alpha1.Service{{
							LoadBalancerSpec: traefikv1alpha1.LoadBalancerSpec{Name: "kotal-dashboard"},
						}, {
							LoadBalancerSpec: traefikv1alpha1.LoadBalancerSpec{Name: "kotal-api"},
						}},
					}},
				},
			}, nil
		}

		ingressRouteUpdateFunc = func(record *traefikv1alpha1.IngressRoute) restErrors.IRestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}
		body, resp := newFiberCtx(validDto, ConfigureDomain, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}
		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)
		assert.EqualValues(t, "something went wrong", result.Message)
	})
	networkIdentifiers = tmpNetworkIdentifiers
	verifyDomainIP = tmpVerifyDomainIp
	verifyDomainHostName = tmpVerifyDomainHostName
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
		settingConfigureRegistrationFunc = func(dto *setting.ConfigureRegistrationRequestDto) restErrors.IRestErr {
			return nil
		}
		body, resp := newFiberCtx(validDto, ConfigureRegistration, locals)

		fmt.Println(string(body), resp)

		var result map[string]pagination.SuccessMessage
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

		assert.Equal(t, badReqErr, result)
		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("configure registration should throw if service throws", func(t *testing.T) {
		settingConfigureRegistrationFunc = func(dto *setting.ConfigureRegistrationRequestDto) restErrors.IRestErr {
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
		settingSettingsFunc = func() ([]*setting.Setting, restErrors.IRestErr) {
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
		settingSettingsFunc = func() ([]*setting.Setting, restErrors.IRestErr) {
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

func TestNetworkIdentifiersAddress(t *testing.T) {
	userDetails := new(token.UserDetails)
	userDetails.ID = "test@test.com"
	var locals = map[string]interface{}{}
	locals["user"] = *userDetails
	t.Run("get ip address should pass", func(t *testing.T) {
		k8svcGetFunc = func(name string, namespace string) (*corev1.Service, restErrors.IRestErr) {
			return &corev1.Service{
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{
							{
								IP:       "1234",
								Hostname: "123.amazon.com",
							},
						},
					},
				},
			}, nil
		}
		body, resp := newFiberCtx("", NetworkIdentifiers, locals)
		fmt.Println(string(body))

		var result map[string]setting.NetworkIdentifierResponseDto
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}
		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
		assert.EqualValues(t, "1234", result["data"].IPAddress)
		assert.EqualValues(t, "123.amazon.com", result["data"].HostName)
	})
	t.Run("get ip address should throw if can't get traefik service", func(t *testing.T) {
		k8svcGetFunc = func(name string, namespace string) (*corev1.Service, restErrors.IRestErr) {
			return nil, restErrors.NewNotFoundError("can't get traefik service")
		}
		body, resp := newFiberCtx("", NetworkIdentifiers, locals)
		fmt.Println(string(body))

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}
		assert.EqualValues(t, http.StatusNotFound, resp.StatusCode)
		assert.EqualValues(t, "can't get traefik service", result.Message)
	})
	t.Run("get ip address should throw if load balancer ip address didn't get provisioned", func(t *testing.T) {
		k8svcGetFunc = func(name string, namespace string) (*corev1.Service, restErrors.IRestErr) {
			return &corev1.Service{
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{},
					},
				},
			}, nil
		}
		body, resp := newFiberCtx("", NetworkIdentifiers, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}
		assert.EqualValues(t, http.StatusNotFound, resp.StatusCode)
		assert.EqualValues(t, http.StatusNotFound, result.Status)
		assert.EqualValues(t, "can't get network identifiers, still provisioning!", result.Message)
	})

}

func TestVerifyDomainHots(t *testing.T) {
	t.Run("verify should throw err with non matching ip address", func(t *testing.T) {
		domain := "example.com"
		ipAddress := "192.168.0.1"
		err := verifyDomainIP(domain, ipAddress)
		assert.EqualValues(t, "Domain DNS records hasn't been updated with an A record that maps example.com to 192.168.0.1.", err.Error())
	})
	t.Run("verify should throw err with invalid domain", func(t *testing.T) {
		domain := "nonexistentdomain123.com"
		ipAddress := "127.0.0.1"
		err := verifyDomainIP(domain, ipAddress)
		assert.NotNil(t, err, "Expected an error")
		assert.Equal(t, "lookup nonexistentdomain123.com: no such host", err.Error(), "Unexpected error message")

	})
}
