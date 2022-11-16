package dbconfig

import (
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"gorm.io/gorm"
)

type service struct{}

type IService interface {
	WithTransaction(txHandle *gorm.DB) IService
	Get(key string) (string, *restErrors.RestErr)
}

var (
	dbConfigRepo = NewRepository()
)

func NewService() IService {
	newService := &service{}
	return newService
}

func (s service) WithTransaction(txHandle *gorm.DB) IService {
	dbConfigRepo = dbConfigRepo.WithTransaction(txHandle)
	return s
}

func (s service) Get(key string) (string, *restErrors.RestErr) {
	return dbConfigRepo.Get(key)
}
