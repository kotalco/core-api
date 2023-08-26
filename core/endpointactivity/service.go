package endpointactivity

import (
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"gorm.io/gorm"
)

type service struct{}

type IService interface {
	WithTransaction(txHandle *gorm.DB) IService
	WithoutTransaction() IService
	GetByEndpointId(endpointId string) (*Activity, restErrors.IRestErr)
	Increment(activity *Activity) restErrors.IRestErr
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

func (s service) GetByEndpointId(endpointId string) (*Activity, restErrors.IRestErr) {
	model, err := activityRepository.FindOne(queryGetByEndpointId, endpointId)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func (s service) Increment(activity *Activity) restErrors.IRestErr {
	activity.Counter++
	return activityRepository.Update(activity)
}
