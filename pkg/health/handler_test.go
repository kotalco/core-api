package health

import (
	"bytes"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var (
	MeasureFunc func(list ...Config) *ResponseDto
)

type HealthMock struct {
	configs map[string]Config
}

// Measure runs all the registered health checks and returns summary status
func (h *HealthMock) Measure(list ...Config) *ResponseDto {
	return MeasureFunc(list...)
}

func TestMain(m *testing.M) {
	h = &HealthMock{}
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

func TestHealth_Healthz(t *testing.T) {
	t.Run("healthz should return ok", func(t *testing.T) {
		MeasureFunc = func(list ...Config) *ResponseDto {
			return &ResponseDto{
				Checks: make([]Check, 0),
				Status: StatusOK,
			}
		}
		body, res := newFiberCtx("", Healthz, map[string]interface{}{})
		var result ResponseDto
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err)
		}
		assert.EqualValues(t, http.StatusOK, res.StatusCode)
	})
	t.Run("healthz should return status unavailable if check faild", func(t *testing.T) {
		MeasureFunc = func(list ...Config) *ResponseDto {
			return &ResponseDto{
				Checks: []Check{
					{
						Name:    "testCheck",
						Status:  StatusUnavailable,
						Failure: "some error",
					},
				},
				Status: StatusUnavailable,
			}
		}
		body, res := newFiberCtx("", Healthz, map[string]interface{}{})
		var result ResponseDto
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err)
		}
		assert.EqualValues(t, http.StatusServiceUnavailable, res.StatusCode)
	})
}
