package endpointactivity

import (
	"github.com/google/uuid"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"gorm.io/gorm"
	"time"
)

type service struct{}

type IService interface {
	WithTransaction(txHandle *gorm.DB) IService
	WithoutTransaction() IService
	Create(endpointId string) restErrors.IRestErr
	MonthlyActivity(endpointId string) (*int, restErrors.IRestErr)
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

func (s service) Create(endpointId string) restErrors.IRestErr {
	record := new(Activity)
	record.ID = uuid.NewString()
	record.EndpointId = endpointId
	record.Timestamp = time.Now().Unix()
	return activityRepository.Create(record)
}

func (s service) MonthlyActivity(endpointId string) (*int, restErrors.IRestErr) {
	list, err := activityRepository.FindMany(queryGetMonthlyActivity, endpointId)
	if err != nil {
		return nil, err
	}
	monthlyCount := len(list)
	return &monthlyCount, nil
}
