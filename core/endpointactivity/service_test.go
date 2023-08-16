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
	GetByEndpointIdFunc    func(endpointId string) (*Activity, restErrors.IRestErr)
	UpdateFunc             func(activity *Activity) restErrors.IRestErr
)

type activityRepoMock struct{}

func (r activityRepoMock) WithTransaction(txHandle *gorm.DB) IRepository {
	return r
}
func (r activityRepoMock) WithoutTransaction() IRepository {
	return r
}

func (r activityRepoMock) GetByEndpointId(endpointId string) (*Activity, restErrors.IRestErr) {
	return GetByEndpointIdFunc(endpointId)
}
func (r activityRepoMock) Update(activity *Activity) restErrors.IRestErr {
	return UpdateFunc(activity)
}

func TestMain(m *testing.M) {
	activityRepository = &activityRepoMock{}
	activityService = NewService()
	code := m.Run()
	os.Exit(code)
}

func TestService_Get_by_Endpoint_Id(t *testing.T) {
	t.Run("Get_activity_by_endpoint_id_should_pass", func(t *testing.T) {

		GetByEndpointIdFunc = func(endpointId string) (*Activity, restErrors.IRestErr) {
			return &Activity{
				ID:         uuid.NewString(),
				EndpointId: "123",
				Counter:    1,
			}, nil
		}

		record, err := activityService.GetByEndpointId("123")
		assert.Nil(t, err)
		assert.NotNil(t, record)
	})
	t.Run("Get_activity_by_endpoint_should_throw_if_service_throws", func(t *testing.T) {
		GetByEndpointIdFunc = func(endpointId string) (*Activity, restErrors.IRestErr) {
			return nil, restErrors.NewNotFoundError("no such record")
		}

		record, err := activityService.GetByEndpointId("123")
		assert.Nil(t, record)
		assert.EqualValues(t, "no such record", err.Error())
	})
}

func TestService_Increment(t *testing.T) {
	t.Run("increment_should_pass", func(t *testing.T) {
		UpdateFunc = func(activity *Activity) restErrors.IRestErr {
			return nil
		}

		restErr := activityService.Increment(new(Activity))
		assert.Nil(t, restErr)
	})

	t.Run("increment_should_throw_if_repos_throws", func(t *testing.T) {
		UpdateFunc = func(activity *Activity) restErrors.IRestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		restErr := activityService.Increment(new(Activity))
		assert.EqualValues(t, "something went wrong", restErr.Error())
	})
}
