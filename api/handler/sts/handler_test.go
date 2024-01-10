package sts

import (
	"bytes"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	restErrors "github.com/kotalco/cloud-api/pkg/errors"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type statefulSetServiceMocks struct{}

var (
	stsCountFunc func() (uint, restErrors.IRestErr)
	stsListFunc  func(namespace string) (appsv1.StatefulSetList, restErrors.IRestErr)
)

func (s statefulSetServiceMocks) Count() (uint, restErrors.IRestErr) {
	return stsCountFunc()
}

func (s statefulSetServiceMocks) List(namespace string) (appsv1.StatefulSetList, restErrors.IRestErr) {
	return stsListFunc(namespace)
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
	statefulSetService = &statefulSetServiceMocks{}
	code := m.Run()
	os.Exit(code)
}

func TestCount(t *testing.T) {
	var locals = map[string]interface{}{}
	locals["namespace"] = "default"

	t.Run("sts count should pass", func(t *testing.T) {
		stsListFunc = func(namespace string) (appsv1.StatefulSetList, restErrors.IRestErr) {
			return appsv1.StatefulSetList{
				Items: []appsv1.StatefulSet{{
					ObjectMeta: v1.ObjectMeta{
						Labels: map[string]string{"kotal.io/protocol": "ipfs"},
					},
				}},
			}, nil
		}
		body, resp := newFiberCtx("", Count, locals)
		var result map[string]map[string]uint
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}
		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
	})
	t.Run("sts should throw if can't list sts service throws", func(t *testing.T) {
		stsListFunc = func(namespace string) (appsv1.StatefulSetList, restErrors.IRestErr) {
			return appsv1.StatefulSetList{}, restErrors.NewInternalServerError("something went wrong")
		}
		_, resp := newFiberCtx("", Count, locals)
		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)

	})

}
