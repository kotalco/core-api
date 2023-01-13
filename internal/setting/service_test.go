package setting

import (
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"os"
	"testing"
)

var (
	settingService    IService
	settingGetFunc    func(key string) (string, *restErrors.RestErr)
	settingCreateFunc func(key string, value string) *restErrors.RestErr
	settingUpdateFunc func(key string, value string) *restErrors.RestErr
	settingFindFunc   func() ([]*Setting, *restErrors.RestErr)
)

type settingRepoMocks struct{}

func (s settingRepoMocks) WithoutTransaction() IRepository {
	return s
}

func (s settingRepoMocks) WithTransaction(txHandle *gorm.DB) IRepository {
	return s
}

func (s settingRepoMocks) Get(key string) (string, *restErrors.RestErr) {
	return settingGetFunc(key)
}

func (s settingRepoMocks) Create(key string, value string) *restErrors.RestErr {
	return settingCreateFunc(key, value)
}

func (s settingRepoMocks) Update(key string, value string) *restErrors.RestErr {
	return settingUpdateFunc(key, value)
}

func (s settingRepoMocks) Find() ([]*Setting, *restErrors.RestErr) {
	return settingFindFunc()
}

func TestMain(m *testing.M) {
	sqlclient.OpenDBConnection()
	settingService = NewService()
	settingRepo = &settingRepoMocks{}
	code := m.Run()
	os.Exit(code)
}

func TestService_Settings(t *testing.T) {
	t.Run("setting should pass", func(t *testing.T) {
		settingFindFunc = func() ([]*Setting, *restErrors.RestErr) {
			return []*Setting{{}}, nil
		}
		record, err := settingService.Settings()
		assert.Nil(t, err)
		assert.NotNil(t, record)
	})

	t.Run("setting should throw if repo throws", func(t *testing.T) {
		settingFindFunc = func() ([]*Setting, *restErrors.RestErr) {
			return nil, restErrors.NewInternalServerError("something went wrong")
		}
		record, err := settingService.Settings()
		assert.Nil(t, record)
		assert.EqualValues(t, "something went wrong", err.Message)
	})
}

func TestService_ConfigureDomain(t *testing.T) {
	t.Run("configure domain should pass with and create new record", func(t *testing.T) {
		settingGetFunc = func(key string) (string, *restErrors.RestErr) {
			return "", restErrors.NewNotFoundError("no such record")
		}
		settingCreateFunc = func(key string, value string) *restErrors.RestErr {
			return nil
		}
		err := settingService.ConfigureDomain(&ConfigureDomainRequestDto{})
		assert.Nil(t, err)
	})
	t.Run("configure domain should pass with and update the old record", func(t *testing.T) {
		settingGetFunc = func(key string) (string, *restErrors.RestErr) {
			return "value", nil
		}
		settingUpdateFunc = func(key string, value string) *restErrors.RestErr {
			return nil
		}
		err := settingService.ConfigureDomain(&ConfigureDomainRequestDto{})
		assert.Nil(t, err)
	})

	t.Run("configure domain should throw if repo throws", func(t *testing.T) {
		settingGetFunc = func(key string) (string, *restErrors.RestErr) {
			return "", restErrors.NewNotFoundError("")
		}
		settingCreateFunc = func(key string, value string) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}
		err := settingService.ConfigureDomain(&ConfigureDomainRequestDto{})
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
func TestService_ConfigureRegistration(t *testing.T) {
	t.Run("configure registration should pass with and create new record", func(t *testing.T) {
		settingGetFunc = func(key string) (string, *restErrors.RestErr) {
			return "", restErrors.NewNotFoundError("no such record")
		}
		settingCreateFunc = func(key string, value string) *restErrors.RestErr {
			return nil
		}
		enableReg := true
		err := settingService.ConfigureRegistration(&ConfigureRegistrationRequestDto{EnableRegistration: &enableReg})
		assert.Nil(t, err)
	})
	t.Run("configure registration should pass with and update the old record", func(t *testing.T) {
		settingGetFunc = func(key string) (string, *restErrors.RestErr) {
			return "value", nil
		}
		settingUpdateFunc = func(key string, value string) *restErrors.RestErr {
			return nil
		}
		enableReg := true
		err := settingService.ConfigureRegistration(&ConfigureRegistrationRequestDto{EnableRegistration: &enableReg})
		assert.Nil(t, err)
	})

	t.Run("configure registration should throw if repo throws", func(t *testing.T) {
		settingGetFunc = func(key string) (string, *restErrors.RestErr) {
			return "", restErrors.NewNotFoundError("")
		}
		settingCreateFunc = func(key string, value string) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}
		enableReg := true
		err := settingService.ConfigureRegistration(&ConfigureRegistrationRequestDto{EnableRegistration: &enableReg})
		assert.EqualValues(t, "something went wrong", err.Message)
	})
}
func TestService_IsRegistrationEnabled(t *testing.T) {
	t.Run("is registration enabled should pass", func(t *testing.T) {
		settingGetFunc = func(key string) (string, *restErrors.RestErr) {
			return "true", nil
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

func TestService_ConfigureActivationKey(t *testing.T) {
	t.Run("configure activation key should pass with and create new record", func(t *testing.T) {
		settingGetFunc = func(key string) (string, *restErrors.RestErr) {
			return "", restErrors.NewNotFoundError("no such record")
		}
		settingCreateFunc = func(key string, value string) *restErrors.RestErr {
			return nil
		}
		err := settingService.ConfigureActivationKey("new key")
		assert.Nil(t, err)
	})
	t.Run("configure activation key  should pass with and update the old record", func(t *testing.T) {
		settingGetFunc = func(key string) (string, *restErrors.RestErr) {
			return "value", nil
		}
		settingUpdateFunc = func(key string, value string) *restErrors.RestErr {
			return nil
		}
		err := settingService.ConfigureActivationKey("key")
		assert.Nil(t, err)
	})

	t.Run("configure activation key should throw if repo throws", func(t *testing.T) {
		settingGetFunc = func(key string) (string, *restErrors.RestErr) {
			return "", restErrors.NewNotFoundError("")
		}
		settingCreateFunc = func(key string, value string) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}
		err := settingService.ConfigureActivationKey("key")
		assert.EqualValues(t, "something went wrong", err.Message)
	})
}

func TestService_GetActivationKey(t *testing.T) {
	t.Run("get activation key  should pass", func(t *testing.T) {
		settingGetFunc = func(key string) (string, *restErrors.RestErr) {
			return "value", nil
		}
		record, err := settingService.GetActivationKey()
		assert.Nil(t, err)
		assert.EqualValues(t, "value", record)
	})
	t.Run("get activation key  should throw", func(t *testing.T) {
		settingGetFunc = func(key string) (string, *restErrors.RestErr) {
			return "", restErrors.NewNotFoundError("no such record")
		}
		_, err := settingService.GetActivationKey()
		assert.EqualValues(t, "no such record", err.Message)
	})
}
