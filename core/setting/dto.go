package setting

import (
	"github.com/go-playground/validator/v10"
	k8svc "github.com/kotalco/cloud-api/pkg/k8s/svc"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
)

const (
	DomainKey       = "domain"
	RegistrationKey = "registration_is_enabled"
	ActivationKey   = "activation_key"
)

type ConfigureDomainRequestDto struct {
	Domain string `json:"domain" validate:"required"`
}

type ConfigureRegistrationRequestDto struct {
	EnableRegistration *bool `json:"enable_registration" validate:"required,boolean"`
}

type SettingResponseDto struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type IPAddressResponseDto struct {
	IPAddress string `json:"ip_address"`
}

func Validate(dto interface{}) restErrors.IRestErr {
	newValidator := validator.New()
	err := newValidator.Struct(dto)

	if err != nil {
		fields := map[string]string{}
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Field() {
			case "Domain":
				fields["domain"] = "invalid domain"
				break
			case "EnableRegistration":
				fields["enable_registration"] = "invalid registration value"
				break
			}
		}

		if len(fields) > 0 {
			return restErrors.NewValidationError(fields)
		}
	}

	return nil
}

// Marshall creates user response from user model
func (dto SettingResponseDto) Marshall(model *Setting) SettingResponseDto {
	dto.Key = model.Key
	dto.Value = model.Value
	return dto
}

// GetDomainBaseUrl get the app baseurl and error if any
// If the user configured his/her domain, if not it gets the traefik external ip
func GetDomainBaseUrl() (string, restErrors.IRestErr) {
	repo := NewRepository()
	url, _ := repo.Get(DomainKey)
	if url != "" {
		return url, nil
	}

	k8service := k8svc.NewService()
	record, err := k8service.Get("traefik", "traefik")
	if err != nil {
		go logger.Error("SEND_GRID_GET_DOMAIN_BASE_URL", err)
		return "", restErrors.NewInternalServerError("can't get traefik service")
	}
	return record.Status.LoadBalancer.Ingress[0].IP, nil
}