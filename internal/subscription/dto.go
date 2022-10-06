package subscription

import (
	"github.com/go-playground/validator/v10"
	restErrors "github.com/kotalco/api/pkg/errors"
)

type AcknowledgementRequestDto struct {
	ActivationKey string `json:"activation_key" validate:"required"`
}

type SubscriptionDto struct {
	ID                     string      `json:"id"`
	Status                 string      `json:"status"`
	Name                   string      `json:"name"`
	ActivationKey          string      `json:"activation_key"`
	StartDate              int64       `json:"start_date"`
	EndDate                int64       `json:"end_date"`
	CanceledAt             int64       `json:"canceled_at"`
	Invoice                interface{} `json:"invoice,omitempty"`
	DefaultPaymentMethodId string      `json:"default_payment_method_id"`
	Attached               bool        `json:"attached"`
}
type LicenseAcknowledgmentDto struct {
	Signature    string          `json:"signature"`
	Subscription SubscriptionDto `json:"subscription"`
}

type CurrentTimeStampDto struct {
	Signature string `json:"signature"`
	Time      struct {
		CurrentTime int64 `json:"current_time"`
	} `json:"time"`
}

func Validate(dto interface{}) *restErrors.RestErr {
	newValidator := validator.New()
	err := newValidator.Struct(dto)

	if err != nil {
		fields := map[string]string{}
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Field() {
			case "ActivationKey":
				fields["activation_key"] = "invalid key"
				break
			}
		}
		if len(fields) > 0 {
			return restErrors.NewValidationError(fields)
		}
	}

	return nil
}
