package svc

import (
	"bytes"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/core-api/core/workspace"
	k8svc "github.com/kotalco/core-api/k8s/svc"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

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
func (s *k8sServiceMock) Create(obj *corev1.Service) restErrors.IRestErr {
	return svcServiceCreateFunc(obj)
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
	svcService = &k8sServiceMock{}
	code := m.Run()
	os.Exit(code)
}

func TestList(t *testing.T) {
	workspaceModel := new(workspace.Workspace)
	var locals = map[string]interface{}{}
	locals["workspace"] = *workspaceModel

	t.Run("list services should pass", func(t *testing.T) {
		k8svcListFunc = func(namespace string) (*corev1.ServiceList, restErrors.IRestErr) {
			return &corev1.ServiceList{
				Items: []corev1.Service{
					{
						Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{{}}},
					},
				},
			}, nil

		}
		availableProtocol = func(protocol string) bool {
			return true
		}

		body, resp := newFiberCtx("", List, locals)
		var result map[string][]k8svc.SvcDto
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)

		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
		assert.EqualValues(t, 1, len(result["data"]))

	})
	t.Run("list services should return empty list if there is no service with valid protocols", func(t *testing.T) {
		k8svcListFunc = func(namespace string) (*corev1.ServiceList, restErrors.IRestErr) {
			return &corev1.ServiceList{
				Items: []corev1.Service{
					{
						Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{{}}},
					},
				},
			}, nil

		}
		availableProtocol = func(protocol string) bool {
			return false
		}

		body, resp := newFiberCtx("", List, locals)
		var result map[string][]k8svc.SvcDto
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)
		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
		assert.EqualValues(t, 0, len(result["data"]))

	})
	t.Run("list services should throw if service.list throws", func(t *testing.T) {
		k8svcListFunc = func(namespace string) (*corev1.ServiceList, restErrors.IRestErr) {
			return nil, restErrors.NewInternalServerError("something went wrong")

		}

		body, resp := newFiberCtx("", List, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		assert.Nil(t, err)

		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)
	})

}
