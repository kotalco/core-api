package onboarding

import (
	"github.com/kotalco/cloud-api/pkg/config"
	"github.com/kotalco/cloud-api/pkg/keystore/dbkeystore"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"gorm.io/gorm"
)

type service struct{}

type IService interface {
	WithTransaction(txHandle *gorm.DB) IService
	SetDomainBaseurl(dto *SetDomainBaseUrlRequestDto) *restErrors.RestErr
	DomainBaseUrlExists() bool
}

var (
	dbKeyStoreService = dbkeystore.NewService()
)

func NewService() IService {
	newService := &service{}
	return newService
}

func (s service) WithTransaction(txHandle *gorm.DB) IService {
	dbKeyStoreService = dbKeyStoreService.WithTransaction(txHandle)
	return s
}

func (s service) SetDomainBaseurl(dto *SetDomainBaseUrlRequestDto) *restErrors.RestErr {
	return dbKeyStoreService.Set(config.Environment.DomainMatchBaseURLKey, dto.DomainBaseUrl)
}

func (s service) DomainBaseUrlExists() bool {
	key, _ := dbKeyStoreService.Get(config.Environment.DomainMatchBaseURLKey)
	if key != "" {
		return true
	}
	return false
}
