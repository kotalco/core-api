package endpointactivity

import (
	"github.com/google/uuid"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"os"
	"testing"
)

var (
	activityService        IService
	WithTransactionFunc    func(txHandle *gorm.DB) IRepository
	WithoutTransactionFunc func() IRepository
	CreateFunc             func(activity *Activity) restErrors.IRestErr
)

type activityRepoMock struct{}

func (r activityRepoMock) WithTransaction(txHandle *gorm.DB) IRepository {
	return r
}
func (r activityRepoMock) WithoutTransaction() IRepository {
	return r
}

func (r activityRepoMock) Create(activity *Activity) restErrors.IRestErr {
	return CreateFunc(activity)
}

func TestMain(m *testing.M) {
	activityRepository = &activityRepoMock{}
	activityService = NewService()
	code := m.Run()
	os.Exit(code)
}

func TestService_Create(t *testing.T) {
	t.Run("create_should_pass", func(t *testing.T) {
		CreateFunc = func(activity *Activity) restErrors.IRestErr {
			return nil
		}

		restErr := activityService.Create(uuid.NewString())
		assert.Nil(t, restErr)
	})

	t.Run("create_should_throw_if_repos_throws", func(t *testing.T) {
		CreateFunc = func(activity *Activity) restErrors.IRestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		restErr := activityService.Create(uuid.NewString())
		assert.EqualValues(t, "something went wrong", restErr.Error())
	})
}
