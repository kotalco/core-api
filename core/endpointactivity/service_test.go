package endpointactivity

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/kotalco/cloud-api/pkg/security"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"os"
	"strings"
	"testing"
)

var (
	activityService        IService
	WithTransactionFunc    func(txHandle *gorm.DB) IRepository
	WithoutTransactionFunc func() IRepository
	CreateInBatchesFunc    func(activities []*Activity) restErrors.IRestErr
	RawQueryFunc           func(query string, dest interface{}, conditions ...interface{}) restErrors.IRestErr
)

type activityRepoMock struct{}

func (r activityRepoMock) WithTransaction(txHandle *gorm.DB) IRepository {
	return r
}
func (r activityRepoMock) WithoutTransaction() IRepository {
	return r
}

func (r activityRepoMock) CreateInBatches(activities []*Activity) restErrors.IRestErr {
	return CreateInBatchesFunc(activities)
}

func (r activityRepoMock) RawQuery(query string, dest interface{}, conditions ...interface{}) restErrors.IRestErr {
	return RawQueryFunc(query, dest, conditions...)
}

func TestMain(m *testing.M) {
	activityRepository = &activityRepoMock{}
	activityService = NewService()
	code := m.Run()
	os.Exit(code)
}

func TestService_Create(t *testing.T) {
	t.Run("create_should_pass", func(t *testing.T) {
		CreateInBatchesFunc = func(activities []*Activity) restErrors.IRestErr {
			return nil
		}

		restErr := activityService.Create([]CreateEndpointActivityDto{{fmt.Sprintf("%s%s", strings.ToLower(security.GenerateRandomString(10)), strings.Replace(uuid.NewString(), "-", "", -1)), 1}})
		assert.Nil(t, restErr)
	})

	t.Run("create_should_throw_if_repos_throws", func(t *testing.T) {
		CreateInBatchesFunc = func(activities []*Activity) restErrors.IRestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		restErr := activityService.Create([]CreateEndpointActivityDto{{fmt.Sprintf("%s%s", strings.ToLower(security.GenerateRandomString(10)), strings.Replace(uuid.NewString(), "-", "", -1)), 1}})
		assert.EqualValues(t, "something went wrong", restErr.Error())
	})
}

func TestService_Stats(t *testing.T) {
	t.Run("stats_should_pass", func(t *testing.T) {
		RawQueryFunc = func(query string, dest interface{}, conditions ...interface{}) restErrors.IRestErr {
			return nil
		}
		activity, restErr := activityService.Stats("123")
		assert.Nil(t, restErr)
		assert.NotNil(t, activity)
	})
	t.Run("stats_should_throw", func(t *testing.T) {
		RawQueryFunc = func(query string, dest interface{}, conditions ...interface{}) restErrors.IRestErr {
			return restErrors.NewInternalServerError("some thing went wrong!")
		}
		activity, restErr := activityService.Stats("123")
		assert.NotNil(t, restErr)
		assert.Nil(t, activity)
	})
}
