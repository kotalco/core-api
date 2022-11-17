package onboarding

import (
	"github.com/go-playground/validator/v10"
	restErrors "github.com/kotalco/community-api/pkg/errors"
)

type SetDomainBaseUrlRequestDto struct {
	DomainBaseUrl string `json:"domain_base_url"`
}

func Validate(dto interface{}) *restErrors.RestErr {
	newValidator := validator.New()
	err := newValidator.Struct(dto)

	if err != nil {
		fields := map[string]string{}
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Field() {
			case "DomainBaseUrl":
				fields["domain_base_url"] = "invalid domain"
				break
			}
		}

		if len(fields) > 0 {
			return restErrors.NewValidationError(fields)
		}
	}

	return nil
}
