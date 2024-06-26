package k8s

import (
	"github.com/go-playground/validator/v10"
	"github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/logger"
	sharedAPI "github.com/kotalco/kotal/apis/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
)

type MetaDataDto struct {
	Name      string `json:"name" validate:"regexp,lt=64"`
	Namespace string `json:"namespace,omitempty"`
}

func (metaDto *MetaDataDto) ObjectMetaFromMetadataDto() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      metaDto.Name,
		Namespace: metaDto.Namespace,
	}
}

func (dto *MetaDataDto) Validate() errors.IRestErr {
	newValidator := validator.New()
	err := newValidator.RegisterValidation("regexp", func(fl validator.FieldLevel) bool {
		re := regexp.MustCompile("^([a-z]|[0-9])([a-z]|[0-9]|-)+([a-z]|[0-9])$")
		return re.MatchString(fl.Field().String())
	})
	if err != nil {
		logger.Warn("USER_DTO_VALIDATE", err)
		return errors.NewInternalServerError("something went wrong!")
	}

	err = newValidator.Struct(dto)

	if err != nil {
		fields := map[string]string{}
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Field() {
			case "Name":
				fields["name"] = "name must start and end with an alphanumeric, and contains no more than 64 alphanumeric characters and - in total."
			}
		}

		if len(fields) > 0 {
			return errors.NewValidationError(fields)
		}
	}

	return nil
}

func DefaultResources(res *sharedAPI.Resources) {
	res.CPU = "1"
	res.Memory = "1Gi"
}
