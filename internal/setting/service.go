package setting

import (
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"gorm.io/gorm"
	"net/http"
)

type service struct{}

type IService interface {
	WithTransaction(txHandle *gorm.DB) IService
	Settings() ([]*Setting, *restErrors.RestErr)
	ConfigureDomain(dto *ConfigureDomainRequestDto) *restErrors.RestErr
	IsDomainConfigured() bool
}

var (
	settingRepo = NewRepository()
)

func NewService() IService {
	newService := &service{}
	return newService
}

func (s service) WithTransaction(txHandle *gorm.DB) IService {
	settingRepo = settingRepo.WithTransaction(txHandle)
	return s
}

func (s service) Settings() ([]*Setting, *restErrors.RestErr) {
	return settingRepo.Find()
}

func (s service) ConfigureDomain(dto *ConfigureDomainRequestDto) *restErrors.RestErr {
	_, err := settingRepo.Get(DomainKey)
	if err != nil {
		if err.Status == http.StatusNotFound {
			//the record doesn't exist create new one
			return settingRepo.Create(DomainKey, dto.Domain)
		}
		return err
	}
	//record exits update it
	return settingRepo.Update(DomainKey, dto.Domain)
}

func (s service) IsDomainConfigured() bool {
	key, _ := settingRepo.Get(DomainKey)
	if key != "" {
		return true
	}
	return false
}
