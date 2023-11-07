package endpointactivity

import (
	"github.com/google/uuid"
	"github.com/kotalco/cloud-api/pkg/config"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"gorm.io/gorm"
	"strconv"
	"time"
)

type service struct{}

type IService interface {
	WithTransaction(txHandle *gorm.DB) IService
	WithoutTransaction() IService
	Create(endpointId string, count int) restErrors.IRestErr
	MonthlyActivity(endpointId string) (int64, restErrors.IRestErr)
	UserMinuteActivity(userId string) (int64, restErrors.IRestErr)
}

var activityRepository = NewRepository()

func NewService() IService {
	newService := &service{}
	return newService
}

func (s service) WithTransaction(txHandle *gorm.DB) IService {
	activityRepository = activityRepository.WithTransaction(txHandle)
	return s
}
func (s service) WithoutTransaction() IService {
	activityRepository = activityRepository.WithoutTransaction()
	return s
}

func (s service) Create(endpointId string, count int) restErrors.IRestErr {
	endpointPortIdLength, err := strconv.Atoi(config.Environment.EndpointPortIdLength)
	if err != nil {
		return restErrors.NewInternalServerError(err.Error())
	}

	parsedUUID, err := uuid.Parse(endpointId[endpointPortIdLength:])
	if err != nil {
		return restErrors.NewInternalServerError(err.Error())
	}

	activities := make([]*Activity, 0)
	for i := 0; i < count; i++ {
		record := new(Activity)
		record.ID = uuid.NewString()
		record.EndpointId = endpointId
		record.UserId = parsedUUID.String()
		record.Timestamp = time.Now().Unix()
		activities = append(activities, record)
	}

	return activityRepository.CreateInBatches(activities)
}

func (s service) MonthlyActivity(endpointId string) (int64, restErrors.IRestErr) {
	count, err := activityRepository.Count(queryGetMonthlyActivity, endpointId)
	if err != nil {
		return 0, err
	}
	return count, nil
}
func (s service) UserMinuteActivity(userId string) (int64, restErrors.IRestErr) {
	count, err := activityRepository.Count(queryGetUserMinuteActivity, userId)
	if err != nil {
		return 0, err
	}
	return count, nil
}
