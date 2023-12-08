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
	Stats(endpointId string) (*ActivityAggregations, restErrors.IRestErr)
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

func (s service) Stats(endpointId string) (*ActivityAggregations, restErrors.IRestErr) {
	activity := &ActivityAggregations{}
	err := activityRepository.RawQuery(rawStatsQuery, activity, endpointId)
	if err != nil {
		return nil, err
	}
	return activity, nil
}
