package dbkeystore

import (
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"gorm.io/gorm"
)

type service struct{}

type IService interface {
	WithTransaction(txHandle *gorm.DB) IService
	Get(key string) (string, *restErrors.RestErr)
	Set(key string, value string) *restErrors.RestErr
}

var (
	keyStoreRpo = NewRepository()
)

func NewService() IService {
	newService := &service{}
	return newService
}

func (s service) WithTransaction(txHandle *gorm.DB) IService {
	keyStoreRpo = keyStoreRpo.WithTransaction(txHandle)
	return s
}

func (s service) Get(key string) (string, *restErrors.RestErr) {
	return keyStoreRpo.Get(key)
}

func (s service) Set(key string, value string) *restErrors.RestErr {
	return keyStoreRpo.Set(key, value)
}
