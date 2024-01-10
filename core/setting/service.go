package setting

import (
	restErrors "github.com/kotalco/cloud-api/pkg/errors"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

type service struct{}

type IService interface {
	WithTransaction(txHandle *gorm.DB) IService
	WithoutTransaction() IService
	Settings() ([]*Setting, restErrors.IRestErr)
	ConfigureDomain(dto *ConfigureDomainRequestDto) restErrors.IRestErr
	IsDomainConfigured() bool
	ConfigureRegistration(dto *ConfigureRegistrationRequestDto) restErrors.IRestErr
	IsRegistrationEnabled() bool
	//ConfigureActivationKey create or update the current subscription activation key
	ConfigureActivationKey(key string) restErrors.IRestErr
	GetActivationKey() (string, restErrors.IRestErr)
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
func (s service) WithoutTransaction() IService {
	settingRepo = settingRepo.WithoutTransaction()
	return s
}

func (s service) Settings() ([]*Setting, restErrors.IRestErr) {
	return settingRepo.Find()
}

func (s service) ConfigureDomain(dto *ConfigureDomainRequestDto) restErrors.IRestErr {
	_, err := settingRepo.Get(DomainKey)
	if err != nil {
		if err.StatusCode() == http.StatusNotFound {
			//the record doesn't exist create new one
			return settingRepo.Create(DomainKey, dto.Domain)
		}
		return err
	}
	//record exits update it
	return settingRepo.Update(DomainKey, dto.Domain)
}

func (s service) IsDomainConfigured() bool {
	value, _ := settingRepo.Get(DomainKey)
	if value != "" {
		return true
	}
	return false
}

func (s service) ConfigureRegistration(dto *ConfigureRegistrationRequestDto) restErrors.IRestErr {
	_, err := settingRepo.Get(RegistrationKey)
	if err != nil {
		if err.StatusCode() == http.StatusNotFound {
			//the record doesn't exist create new one
			return settingRepo.Create(RegistrationKey, strconv.FormatBool(*dto.EnableRegistration))
		}
		return err
	}
	//record exits update it
	return settingRepo.Update(RegistrationKey, strconv.FormatBool(*dto.EnableRegistration))
}

func (s service) IsRegistrationEnabled() bool {
	value, _ := settingRepo.Get(RegistrationKey)
	if value == "true" {
		return true
	}
	return false
}

func (s service) ConfigureActivationKey(key string) restErrors.IRestErr {
	_, err := settingRepo.Get(ActivationKey)
	if err != nil {
		if err.StatusCode() == http.StatusNotFound {
			//the record doesn't exist create new one
			return settingRepo.Create(ActivationKey, key)
		}
		return err
	}
	//record exits update it
	return settingRepo.Update(ActivationKey, key)
}

func (s service) GetActivationKey() (string, restErrors.IRestErr) {
	return settingRepo.Get(ActivationKey)
}
