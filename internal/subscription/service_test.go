package subscription

import (
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	subscriptionAPI "github.com/kotalco/cloud-api/pkg/subscription"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"os"
	"testing"
	"time"
)

/*
subscriptionAPI service  mocks
*/
var (
	subscriptionAPIAcknowledgmentFunc   func(activationKey string, clusterID string) ([]byte, *restErrors.RestErr)
	subscriptionAPICurrentTimeStampFunc func() ([]byte, *restErrors.RestErr)
	subscriptionService                 IService
)

type subscriptionApiServiceMock struct{}

func (s subscriptionApiServiceMock) Acknowledgment(activationKey string, clusterID string) ([]byte, *restErrors.RestErr) {
	return subscriptionAPIAcknowledgmentFunc(activationKey, clusterID)
}
func (s subscriptionApiServiceMock) CurrentTimeStamp() ([]byte, *restErrors.RestErr) {
	return subscriptionAPICurrentTimeStampFunc()
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

/*
Namespace service Mocks
*/

var (
	namespaceCreateNamespaceFunc func(name string) *restErrors.RestErr
	namespaceGetNamespaceFunc    func(name string) (*corev1.Namespace, *restErrors.RestErr)
	namespaceDeleteNamespaceFunc func(name string) *restErrors.RestErr
)

type namespaceServiceMock struct{}

func (namespaceServiceMock) Create(name string) *restErrors.RestErr {
	return namespaceCreateNamespaceFunc(name)
}

func (namespaceServiceMock) Get(name string) (*corev1.Namespace, *restErrors.RestErr) {
	return namespaceGetNamespaceFunc(name)
}

func (namespaceServiceMock) Delete(name string) *restErrors.RestErr {
	return namespaceDeleteNamespaceFunc(name)
}

func TestMain(m *testing.M) {
	sqlclient.OpenDBConnection()
	//should be called before the mocks coz the mocks should replace the services initiated in the NewService func
	subscriptionService = NewService()
	subscriptionAPIService = &subscriptionApiServiceMock{}
	ecService = &eccServiceMock{}
	namespaceService = &namespaceServiceMock{}

	code := m.Run()
	os.Exit(code)
}

func TestService_Acknowledgment(t *testing.T) {
	t.Run("test service acknowledgment should pass", func(t *testing.T) {
		namespaceGetNamespaceFunc = func(name string) (*corev1.Namespace, *restErrors.RestErr) {
			return &corev1.Namespace{}, nil
		}
		subscriptionAPIAcknowledgmentFunc = func(activationKey string, clusterID string) ([]byte, *restErrors.RestErr) {
			responseBody, _ := json.Marshal(map[string]LicenseAcknowledgmentDto{"data": {Subscription: SubscriptionDto{}}})
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

		subscriptionAPICurrentTimeStampFunc = func() ([]byte, *restErrors.RestErr) {
			requestBody := map[string]CurrentTimeStampDto{
				"data": {
					Signature: base64.StdEncoding.EncodeToString([]byte("1234")),
					Time: struct {
						CurrentTime int64 `json:"current_time"`
					}(struct{ CurrentTime int64 }{CurrentTime: time.Now().Unix()}),
				},
			}
			jsonBody, _ := json.Marshal(requestBody)

			return jsonBody, nil
		}

		err := subscriptionService.Acknowledgment("key")
		assert.Nil(t, err)

	})

	t.Run("test service acknowledgment should throw if can't get cluster id", func(t *testing.T) {
		namespaceGetNamespaceFunc = func(name string) (*corev1.Namespace, *restErrors.RestErr) {
			return nil, restErrors.NewInternalServerError("something went wrong")
		}
		err := subscriptionService.Acknowledgment("")
		assert.EqualValues(t, "can't get cluster details", err.Message)
	})
	t.Run("test service acknowledgment should throw if can't get acknowledgment from subscription api", func(t *testing.T) {
		namespaceGetNamespaceFunc = func(name string) (*corev1.Namespace, *restErrors.RestErr) {
			return &corev1.Namespace{}, nil
		}
		subscriptionAPIAcknowledgmentFunc = func(activationKey string, clusterID string) ([]byte, *restErrors.RestErr) {
			return nil, restErrors.NewInternalServerError("something went wrong")
		}

		err := subscriptionService.Acknowledgment("key")
		assert.EqualValues(t, http.StatusInternalServerError, err.Status)
	})
	t.Run("test service acknowledgment should throw if can't decode public key", func(t *testing.T) {
		namespaceGetNamespaceFunc = func(name string) (*corev1.Namespace, *restErrors.RestErr) {
			return &corev1.Namespace{}, nil
		}
		subscriptionAPIAcknowledgmentFunc = func(activationKey string, clusterID string) ([]byte, *restErrors.RestErr) {
			responseBody, _ := json.Marshal(map[string]LicenseAcknowledgmentDto{"data": {Subscription: SubscriptionDto{}}})
			return responseBody, nil
		}
		eccDecodePublicFunc = func(hexEncodedPub string) (*ecdsa.PublicKey, error) {
			return nil, restErrors.NewInternalServerError("something went wrong")
		}

		err := subscriptionService.Acknowledgment("key")
		assert.EqualValues(t, "can't activate subscription", err.Message)
	})
	t.Run("test service acknowledgment should throw if can't verify signature", func(t *testing.T) {
		namespaceGetNamespaceFunc = func(name string) (*corev1.Namespace, *restErrors.RestErr) {
			return &corev1.Namespace{}, nil
		}
		subscriptionAPIAcknowledgmentFunc = func(activationKey string, clusterID string) ([]byte, *restErrors.RestErr) {
			responseBody, _ := json.Marshal(map[string]LicenseAcknowledgmentDto{"data": {Subscription: SubscriptionDto{}}})
			return responseBody, nil
		}
		eccDecodePublicFunc = func(hexEncodedPub string) (*ecdsa.PublicKey, error) {
			return &ecdsa.PublicKey{}, nil
		}

		eccVerifySignatureFunc = func(data []byte, signature []byte, pubKey *ecdsa.PublicKey) (bool, error) {
			return false, restErrors.NewInternalServerError("something went wrong")
		}

		err := subscriptionService.Acknowledgment("key")
		assert.EqualValues(t, "can't activate subscription", err.Message)
	})
	t.Run("test service acknowledgment should throw if subscription is invalid", func(t *testing.T) {
		namespaceGetNamespaceFunc = func(name string) (*corev1.Namespace, *restErrors.RestErr) {
			return &corev1.Namespace{}, nil
		}
		subscriptionAPIAcknowledgmentFunc = func(activationKey string, clusterID string) ([]byte, *restErrors.RestErr) {
			responseBody, _ := json.Marshal(map[string]LicenseAcknowledgmentDto{"data": {Subscription: SubscriptionDto{}}})
			return responseBody, nil
		}
		eccDecodePublicFunc = func(hexEncodedPub string) (*ecdsa.PublicKey, error) {
			return &ecdsa.PublicKey{}, nil
		}

		eccVerifySignatureFunc = func(data []byte, signature []byte, pubKey *ecdsa.PublicKey) (bool, error) {
			return false, nil
		}

		err := subscriptionService.Acknowledgment("key")
		assert.EqualValues(t, "can't activate subscription", err.Message)
	})

}

func TestService_CurrentTimestamp(t *testing.T) {
	t.Run("current timestamp should pass", func(t *testing.T) {
		currentTime := time.Now().Unix()
		subscriptionAPICurrentTimeStampFunc = func() ([]byte, *restErrors.RestErr) {
			requestBody := map[string]CurrentTimeStampDto{
				"data": {
					Signature: base64.StdEncoding.EncodeToString([]byte("1234")),
					Time: struct {
						CurrentTime int64 `json:"current_time"`
					}(struct{ CurrentTime int64 }{CurrentTime: currentTime}),
				},
			}
			jsonBody, _ := json.Marshal(requestBody)

			return jsonBody, nil
		}

		eccDecodePublicFunc = func(hexEncodedPub string) (*ecdsa.PublicKey, error) {
			return &ecdsa.PublicKey{}, nil
		}

		eccVerifySignatureFunc = func(data []byte, signature []byte, pubKey *ecdsa.PublicKey) (bool, error) {
			return true, nil
		}

		time, err := subscriptionService.CurrentTimestamp()
		assert.Nil(t, err)
		assert.EqualValues(t, currentTime, time)
	})
	t.Run("current timestamp should throw if subscriptionApi.currentTimeStamp throws", func(t *testing.T) {
		subscriptionAPICurrentTimeStampFunc = func() ([]byte, *restErrors.RestErr) {
			return nil, restErrors.NewInternalServerError("something went wrong")
		}

		time, err := subscriptionService.CurrentTimestamp()
		assert.EqualValues(t, http.StatusInternalServerError, err.Status)
		assert.EqualValues(t, 0, time)
	})

	t.Run("current timestamp should throw if can't decode public", func(t *testing.T) {
		currentTime := time.Now().Unix()
		subscriptionAPICurrentTimeStampFunc = func() ([]byte, *restErrors.RestErr) {
			requestBody := map[string]CurrentTimeStampDto{
				"data": {
					Signature: base64.StdEncoding.EncodeToString([]byte("1234")),
					Time: struct {
						CurrentTime int64 `json:"current_time"`
					}(struct{ CurrentTime int64 }{CurrentTime: currentTime}),
				},
			}
			jsonBody, _ := json.Marshal(requestBody)

			return jsonBody, nil
		}

		eccDecodePublicFunc = func(hexEncodedPub string) (*ecdsa.PublicKey, error) {
			return nil, restErrors.NewInternalServerError("something")
		}

		time, err := subscriptionService.CurrentTimestamp()
		assert.EqualValues(t, http.StatusInternalServerError, err.Status)
		assert.EqualValues(t, 0, time)
	})
	t.Run("current timestamp should throw if  verify signature throws internal", func(t *testing.T) {
		currentTime := time.Now().Unix()
		subscriptionAPICurrentTimeStampFunc = func() ([]byte, *restErrors.RestErr) {
			requestBody := map[string]CurrentTimeStampDto{
				"data": {
					Signature: base64.StdEncoding.EncodeToString([]byte("1234")),
					Time: struct {
						CurrentTime int64 `json:"current_time"`
					}(struct{ CurrentTime int64 }{CurrentTime: currentTime}),
				},
			}
			jsonBody, _ := json.Marshal(requestBody)

			return jsonBody, nil
		}

		eccDecodePublicFunc = func(hexEncodedPub string) (*ecdsa.PublicKey, error) {
			return &ecdsa.PublicKey{}, nil
		}

		eccVerifySignatureFunc = func(data []byte, signature []byte, pubKey *ecdsa.PublicKey) (bool, error) {
			return true, restErrors.NewInternalServerError("something went wrong")
		}

		time, err := subscriptionService.CurrentTimestamp()
		assert.EqualValues(t, http.StatusInternalServerError, err.Status)
		assert.EqualValues(t, 0, time)
	})
	t.Run("current timestamp should throw if can't verify signature ", func(t *testing.T) {
		currentTime := time.Now().Unix()
		subscriptionAPICurrentTimeStampFunc = func() ([]byte, *restErrors.RestErr) {
			requestBody := map[string]CurrentTimeStampDto{
				"data": {
					Signature: base64.StdEncoding.EncodeToString([]byte("1234")),
					Time: struct {
						CurrentTime int64 `json:"current_time"`
					}(struct{ CurrentTime int64 }{CurrentTime: currentTime}),
				},
			}
			jsonBody, _ := json.Marshal(requestBody)

			return jsonBody, nil
		}

		eccDecodePublicFunc = func(hexEncodedPub string) (*ecdsa.PublicKey, error) {
			return &ecdsa.PublicKey{}, nil
		}

		eccVerifySignatureFunc = func(data []byte, signature []byte, pubKey *ecdsa.PublicKey) (bool, error) {
			return false, nil
		}

		time, err := subscriptionService.CurrentTimestamp()
		assert.EqualValues(t, http.StatusInternalServerError, err.Status)
		assert.EqualValues(t, 0, time)
	})

}
