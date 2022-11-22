package setting

import (
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"gorm.io/gorm"
)

type service struct{}

type IService interface {
	WithTransaction(txHandle *gorm.DB) IService
	Settings() (*Setting, *restErrors.RestErr)
	ConfigureDomain(dto *ConfigureDomainRequestDto) *restErrors.RestErr
	UpdateDomainConfiguration(dto *ConfigureDomainRequestDto) *restErrors.RestErr
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

func (s service) Settings() (*Setting, *restErrors.RestErr) {
	return settingRepo.Settings()
}

func (s service) ConfigureDomain(dto *ConfigureDomainRequestDto) *restErrors.RestErr {
	return settingRepo.Set(DomainKey, dto.Domain)
}

func (s service) UpdateDomainConfiguration(dto *ConfigureDomainRequestDto) *restErrors.RestErr {
	return settingRepo.Update(DomainKey, dto.Domain)
}

func (s service) IsDomainConfigured() bool {
	key, _ := settingRepo.Get(DomainKey)
	if key != "" {
		return true
	}
	return false
}
