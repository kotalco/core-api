package endpointactivity

import (
	"github.com/google/uuid"
	"github.com/kotalco/core-api/config"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"gorm.io/gorm"
	"strconv"
	"time"
)

type service struct{}

type IService interface {
	WithTransaction(txHandle *gorm.DB) IService
	WithoutTransaction() IService
	Create([]CreateEndpointActivityDto) restErrors.IRestErr
	Stats(startDate time.Time, endDate time.Time, endpointId string) (*[]ActivityAggregations, restErrors.IRestErr)
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

func (s service) Create(activitiesDto []CreateEndpointActivityDto) restErrors.IRestErr {
	endpointPortIdLength, err := strconv.Atoi(config.Environment.EndpointPortIdLength)
	if err != nil {
		return restErrors.NewInternalServerError(err.Error())
	}

	activities := make([]*Activity, 0)
	for _, v := range activitiesDto {
		parsedUUID, err := uuid.Parse(v.RequestId[endpointPortIdLength:])
		if err != nil {
			return restErrors.NewInternalServerError(err.Error())
		}

		for i := 0; i < v.Count; i++ {
			record := new(Activity)
			record.ID = uuid.NewString()
			record.EndpointId = v.RequestId
			record.UserId = parsedUUID.String()
			record.Timestamp = time.Now()
			activities = append(activities, record)
		}
	}

	return activityRepository.CreateInBatches(activities)
}

func (s service) Stats(startDate time.Time, endDate time.Time, endpointId string) (*[]ActivityAggregations, restErrors.IRestErr) {
	activityDest := new([]ActivityAggregations)
	err := activityRepository.RawQuery(activityBetweenDates, activityDest, endpointId, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return activityDest, nil
}
