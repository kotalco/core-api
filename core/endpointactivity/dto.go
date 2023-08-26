package endpointactivity

import (
	"github.com/go-playground/validator/v10"
	restErrors "github.com/kotalco/community-api/pkg/errors"
)

type EndpointActivityDto struct {
	RequestId string `json:"request_id" validate:"required"`
}

func Validate(dto interface{}) restErrors.IRestErr {
	newValidator := validator.New()
	err := newValidator.Struct(dto)

	if err != nil {
		fields := map[string]string{}
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Field() {
			case "RequestId":
				fields["request_id"] = "invalid request id"
				break
			}
		}
		if len(fields) > 0 {
			return restErrors.NewValidationError(fields)
		}
	}

	return nil
}