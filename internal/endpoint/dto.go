package endpoint

import (
	"github.com/go-playground/validator/v10"
	restErrors "github.com/kotalco/community-api/pkg/errors"
)

type CreateEndpointDto struct {
	Name      string `json:"name" validate:"required"`
	Namespace string `json:"namespace" validate:"required"` //todo change to workspace and handle errors
	//Match       string `json:"match"`
	ServiceName string `json:"service_name" validate:"required"`
	ServicePort int    `json:"service_port" validate:"required"`
}

func Validate(dto interface{}) *restErrors.RestErr {
	newValidator := validator.New()
	err := newValidator.Struct(dto)

	if err != nil {
		fields := map[string]string{}
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Field() {
			case "Name":
				fields["name"] = "invalid name"
				break
			case "Namespace":
				fields["namespace"] = "invalid namespace"
				break
			case "Match":
				fields["match"] = "invalid match field"
				break
			case "ServiceName":
				fields["service_name"] = "invalid service_name"
				break
			case "ServicePort":
				fields["service_port"] = "invalid service_port"
				break
			}
		}
		if len(fields) > 0 {
			return restErrors.NewValidationError(fields)
		}
	}

	return nil
}
