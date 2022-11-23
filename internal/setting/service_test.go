package setting

import (
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"os"
	"testing"
)

var (
	settingService      IService
	settingGetFunc      func(key string) (string, *restErrors.RestErr)
	settingSetFunc      func(key string, value string) *restErrors.RestErr
	settingUpdateFunc   func(key string, value string) *restErrors.RestErr
	settingSettingsFunc func() (*Setting, *restErrors.RestErr)
)

type settingRepoMocks struct{}

func (s settingRepoMocks) WithTransaction(txHandle *gorm.DB) IRepository {
	return s
}

func (s settingRepoMocks) Get(key string) (string, *restErrors.RestErr) {
	return settingGetFunc(key)
}

func (s settingRepoMocks) Set(key string, value string) *restErrors.RestErr {
	return settingSetFunc(key, value)
}

func (s settingRepoMocks) Update(key string, value string) *restErrors.RestErr {
	return settingUpdateFunc(key, value)
}

func (s settingRepoMocks) Settings() (*Setting, *restErrors.RestErr) {
	return settingSettingsFunc()
}

func TestMain(m *testing.M) {
	settingService = NewService()
	settingRepo = &settingRepoMocks{}
	code := m.Run()
	os.Exit(code)
}

func TestService_Settings(t *testing.T) {
	t.Run("setting should pass", func(t *testing.T) {
		settingSettingsFunc = func() (*Setting, *restErrors.RestErr) {
			return &Setting{}, nil
		}
		record, err := settingService.Settings()
		assert.Nil(t, err)
		assert.NotNil(t, record)
	})

	t.Run("setting should throw if repo throws", func(t *testing.T) {
		settingSettingsFunc = func() (*Setting, *restErrors.RestErr) {
			return nil, restErrors.NewInternalServerError("something went wrong")
		}
		record, err := settingService.Settings()
		assert.Nil(t, record)
		assert.EqualValues(t, "something went wrong", err.Message)
	})
}

func TestService_ConfigureDomain(t *testing.T) {
	t.Run("configure domain should pass", func(t *testing.T) {
		settingSetFunc = func(key string, value string) *restErrors.RestErr {
			return nil
		}
		err := settingService.ConfigureDomain(&ConfigureDomainRequestDto{})
		assert.Nil(t, err)
	})

	t.Run("configure domain should throw if repo throws", func(t *testing.T) {
		settingSetFunc = func(key string, value string) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}
		err := settingService.ConfigureDomain(&ConfigureDomainRequestDto{})
		assert.EqualValues(t, "something went wrong", err.Message)
	})
}

func TestService_UpdateDomainConfiguration(t *testing.T) {
	t.Run("update domain configuration should pass", func(t *testing.T) {
		settingUpdateFunc = func(key string, value string) *restErrors.RestErr {
			return nil
		}
		err := settingService.UpdateDomainConfiguration(&ConfigureDomainRequestDto{})
		assert.Nil(t, err)
	})

	t.Run("update domain configuration should throw if service throws", func(t *testing.T) {
		settingUpdateFunc = func(key string, value string) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		err := settingService.UpdateDomainConfiguration(&ConfigureDomainRequestDto{})
		assert.EqualValues(t, "something went wrong", err.Message)
	})
}

func TestService_IsDomainConfigured(t *testing.T) {
	t.Run("is domain configured should pass", func(t *testing.T) {
		settingGetFunc = func(key string) (string, *restErrors.RestErr) {
			return "key", nil
		}
		assert.True(t, settingService.IsDomainConfigured())
	})

	t.Run("is domain configured should return false", func(t *testing.T) {
		settingGetFunc = func(key string) (string, *restErrors.RestErr) {
			return "", restErrors.NewInternalServerError("something went wrong")
		}
		assert.False(t, settingService.IsDomainConfigured())
	})
}
