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
	CreateFunc             func(activity *Activity) restErrors.IRestErr
	FindManyFunc           func(query interface{}, conditions ...interface{}) ([]*Activity, restErrors.IRestErr)
	CountFunc              func(query interface{}, conditions ...interface{}) (int64, restErrors.IRestErr)
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

func (r activityRepoMock) FindMany(query interface{}, conditions ...interface{}) ([]*Activity, restErrors.IRestErr) {
	return FindManyFunc(query, conditions)
}
func (r activityRepoMock) Count(query interface{}, conditions ...interface{}) (int64, restErrors.IRestErr) {
	return CountFunc(query, conditions)
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

		restErr := activityService.Create(fmt.Sprintf("%s%s", strings.ToLower(security.GenerateRandomString(10)), strings.Replace(uuid.NewString(), "-", "", -1)))
		assert.Nil(t, restErr)
	})

	t.Run("create_should_throw_if_repos_throws", func(t *testing.T) {
		CreateFunc = func(activity *Activity) restErrors.IRestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		restErr := activityService.Create(fmt.Sprintf("%s%s", strings.ToLower(security.GenerateRandomString(10)), strings.Replace(uuid.NewString(), "-", "", -1)))
		assert.EqualValues(t, "something went wrong", restErr.Error())
	})
}

func TestService_MonthlyActivity(t *testing.T) {
	t.Run("monthly activity should pass", func(t *testing.T) {
		CountFunc = func(query interface{}, conditions ...interface{}) (int64, restErrors.IRestErr) {
			return 1, nil
		}
		count, err := activityService.MonthlyActivity("123")
		assert.Nil(t, err)
		assert.EqualValues(t, 1, count)
	})

	t.Run(" monthly activity should throw if repo throws", func(t *testing.T) {
		CountFunc = func(query interface{}, conditions ...interface{}) (int64, restErrors.IRestErr) {
			return 0, restErrors.NewInternalServerError("something went wrong")
		}
		_, err := activityService.MonthlyActivity("123")

		assert.EqualValues(t, "something went wrong", err.Error())
	})
}
