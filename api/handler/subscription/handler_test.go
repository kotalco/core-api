package subscription

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/api/pkg/shared"
	"github.com/kotalco/cloud-api/internal/subscription"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	subscriptionAPI "github.com/kotalco/cloud-api/pkg/subscription"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var (
	subscriptionAcknowledgmentFunc func(activationKey string) ([]byte, *restErrors.RestErr)
)

type subscriptionApiServiceMock struct{}

func (s subscriptionApiServiceMock) Acknowledgment(activationKey string) ([]byte, *restErrors.RestErr) {
	return subscriptionAcknowledgmentFunc(activationKey)
}

/*
ecc service  mocks
*/
var (
	eccGenerateKeysFunc    func() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error)
	eccEncodePrivateFunc   func(privKey *ecdsa.PrivateKey) (string, error)
	eccEncodePublicFunc    func(pubKey *ecdsa.PublicKey) (string, error)
	eccDecodePrivateFunc   func(hexEncodedPriv string) (*ecdsa.PrivateKey, error)
	eccDecodePublicFunc    func(hexEncodedPub string) (*ecdsa.PublicKey, error)
	eccVerifySignatureFunc func(data []byte, signature []byte, pubKey *ecdsa.PublicKey) (bool, error)
	eccCreateSignatureFunc func(data []byte, privKey *ecdsa.PrivateKey) ([]byte, error)
)

type eccServiceMock struct{}

func (e eccServiceMock) GenerateKeys() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	return eccGenerateKeysFunc()
}

func (e eccServiceMock) EncodePrivate(privKey *ecdsa.PrivateKey) (string, error) {
	return eccEncodePrivateFunc(privKey)
}

func (e eccServiceMock) EncodePublic(pubKey *ecdsa.PublicKey) (string, error) {
	return eccEncodePublicFunc(pubKey)
}

func (e eccServiceMock) DecodePrivate(hexEncodedPriv string) (*ecdsa.PrivateKey, error) {
	return eccDecodePrivateFunc(hexEncodedPriv)
}

func (e eccServiceMock) DecodePublic(hexEncodedPub string) (*ecdsa.PublicKey, error) {
	return eccDecodePublicFunc(hexEncodedPub)
}

func (e eccServiceMock) VerifySignature(data []byte, signature []byte, pubKey *ecdsa.PublicKey) (bool, error) {
	return eccVerifySignatureFunc(data, signature, pubKey)
}

func (e eccServiceMock) CreateSignature(data []byte, privKey *ecdsa.PrivateKey) ([]byte, error) {
	return eccCreateSignatureFunc(data, privKey)
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
	sqlclient.OpenDBConnection()
	subscriptionAPIService = &subscriptionApiServiceMock{}
	ecService = &eccServiceMock{}

	code := m.Run()
	os.Exit(code)
}

func TestAcknowledgement(t *testing.T) {
	var validDto = map[string]string{
		"activation_key": "1234",
	}
	var invalidDto = map[string]string{}
	t.Run("Acknowledgement should pass", func(t *testing.T) {
		subscriptionAcknowledgmentFunc = func(activationKey string) ([]byte, *restErrors.RestErr) {
			responseBody, _ := json.Marshal(map[string]subscription.LicenseAcknowledgmentDto{"data": {Subscription: subscription.SubscriptionDto{}}})
			return responseBody, nil
		}

		eccDecodePublicFunc = func(hexEncodedPub string) (*ecdsa.PublicKey, error) {
			return &ecdsa.PublicKey{}, nil
		}

		eccVerifySignatureFunc = func(data []byte, signature []byte, pubKey *ecdsa.PublicKey) (bool, error) {
			return true, nil
		}
		subscriptionAPI.IsValid = func() bool {
			return true
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
	t.Run("Acknowledgement should throw if subscription-api returns an error", func(t *testing.T) {
		subscriptionAcknowledgmentFunc = func(activationKey string) ([]byte, *restErrors.RestErr) {
			return nil, restErrors.NewInternalServerError("something went wrong")
		}

		body, resp := newFiberCtx(validDto, Acknowledgement, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)

	})
	t.Run("Acknowledgement should throw if can't unmarshall subscription-api response", func(t *testing.T) {
		subscriptionAcknowledgmentFunc = func(activationKey string) ([]byte, *restErrors.RestErr) {
			return []byte(""), nil
		}

		body, resp := newFiberCtx(validDto, Acknowledgement, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)
	})
	t.Run("Acknowledgement should throw if can't decode public key", func(t *testing.T) {
		subscriptionAcknowledgmentFunc = func(activationKey string) ([]byte, *restErrors.RestErr) {
			responseBody, _ := json.Marshal(map[string]subscription.LicenseAcknowledgmentDto{"data": {Subscription: subscription.SubscriptionDto{}}})
			return responseBody, nil
		}

		eccDecodePublicFunc = func(hexEncodedPub string) (*ecdsa.PublicKey, error) {
			return nil, restErrors.NewInternalServerError("something went wrong")
		}

		body, resp := newFiberCtx(validDto, Acknowledgement, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)

	})
	t.Run("Acknowledgement should throw if can't verify signature", func(t *testing.T) {
		subscriptionAcknowledgmentFunc = func(activationKey string) ([]byte, *restErrors.RestErr) {
			responseBody, _ := json.Marshal(map[string]subscription.LicenseAcknowledgmentDto{"data": {Subscription: subscription.SubscriptionDto{}}})
			return responseBody, nil
		}

		eccDecodePublicFunc = func(hexEncodedPub string) (*ecdsa.PublicKey, error) {
			return &ecdsa.PublicKey{}, nil
		}

		eccVerifySignatureFunc = func(data []byte, signature []byte, pubKey *ecdsa.PublicKey) (bool, error) {
			return false, restErrors.NewInternalServerError("something went wrong")
		}

		body, resp := newFiberCtx(validDto, Acknowledgement, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)

	})
	t.Run("Acknowledgement should throw if signature invalid", func(t *testing.T) {
		subscriptionAcknowledgmentFunc = func(activationKey string) ([]byte, *restErrors.RestErr) {
			responseBody, _ := json.Marshal(map[string]subscription.LicenseAcknowledgmentDto{"data": {Subscription: subscription.SubscriptionDto{}}})
			return responseBody, nil
		}

		eccDecodePublicFunc = func(hexEncodedPub string) (*ecdsa.PublicKey, error) {
			return &ecdsa.PublicKey{}, nil
		}

		eccVerifySignatureFunc = func(data []byte, signature []byte, pubKey *ecdsa.PublicKey) (bool, error) {
			return false, nil
		}

		body, resp := newFiberCtx(validDto, Acknowledgement, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)

	})
	t.Run("Acknowledgement should throw if subscription is invalid", func(t *testing.T) {
		subscriptionAcknowledgmentFunc = func(activationKey string) ([]byte, *restErrors.RestErr) {
			responseBody, _ := json.Marshal(map[string]subscription.LicenseAcknowledgmentDto{"data": {Subscription: subscription.SubscriptionDto{}}})
			return responseBody, nil
		}

		eccDecodePublicFunc = func(hexEncodedPub string) (*ecdsa.PublicKey, error) {
			return &ecdsa.PublicKey{}, nil
		}

		eccVerifySignatureFunc = func(data []byte, signature []byte, pubKey *ecdsa.PublicKey) (bool, error) {
			return true, nil
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

		assert.EqualValues(t, http.StatusGone, resp.StatusCode)

	})
}
