package endpointactivity

import (
	"github.com/go-playground/validator/v10"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"regexp"
	"time"
)

type ActivityAggregations struct {
	Date     time.Time `json:"date"`
	Activity int       `json:"activity"`
}

type CreateEndpointActivityDto struct {
	RequestId string `json:"request_id" validate:"required"`
	Count     int    `json:"count" validate:"required"`
}

const (
	LastMonth = "last_month"
	LastWeek  = "last_week"
)

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

func GetEndpointId(path string) string {
	// Compile the regular expression
	re := regexp.MustCompile("([a-z0-9]{42})")
	// Find the first match of the pattern in the URL Path
	match := re.FindStringSubmatch(path)

	if len(match) == 0 {
		return ""
	}
	return match[0]
}
