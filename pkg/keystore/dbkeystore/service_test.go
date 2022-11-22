package dbkeystore

import (
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"os"
	"testing"
)

var (
	keyStoreService             IService
	keystoreWithTransactionFunc func(txHandle *gorm.DB) IRepository
	keystoreGetFunc             func(key string) (string, *restErrors.RestErr)
	keystoreSetFunc             func(key string, value string) *restErrors.RestErr
)

type keystoreRepoMocks struct{}

func (k keystoreRepoMocks) WithTransaction(txHandle *gorm.DB) IRepository {
	return k
}

func (k keystoreRepoMocks) Get(key string) (string, *restErrors.RestErr) {
	return keystoreGetFunc(key)
}

func (k keystoreRepoMocks) Set(key string, value string) *restErrors.RestErr {
	return keystoreSetFunc(key, value)
}

func TestMain(m *testing.M) {
	keyStoreRpo = &keystoreRepoMocks{}
	keyStoreService = NewService()
	code := m.Run()
	os.Exit(code)
}

func TestService_Get(t *testing.T) {
	t.Run("get key store should pass", func(t *testing.T) {
		keystoreGetFunc = func(key string) (string, *restErrors.RestErr) {
			return "value", nil
		}
		value, err := keyStoreService.Get("")
		assert.Nil(t, err)
		assert.EqualValues(t, "value", value)
	})

	t.Run("get key store should throw if repo throws", func(t *testing.T) {
		keystoreGetFunc = func(key string) (string, *restErrors.RestErr) {
			return "", restErrors.NewNotFoundError("no such record")
		}
		value, err := keyStoreService.Get("")
		assert.EqualValues(t, "", value)
		assert.EqualValues(t, "no such record", err.Message)
	})
}

func TestService_Set(t *testing.T) {
	t.Run("set key store should pass", func(t *testing.T) {
		keystoreSetFunc = func(key string, value string) *restErrors.RestErr {
			return nil
		}

		err := keyStoreService.Set("", "")
		assert.Nil(t, err)
	})

	t.Run("set key store should throw if service throws", func(t *testing.T) {
		keystoreSetFunc = func(key string, value string) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}
		err := keyStoreService.Set("", "")
		assert.EqualValues(t, "something went wrong", err.Message)
	})
}
