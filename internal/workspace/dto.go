package workspace

import (
	"github.com/go-playground/validator/v10"
	restErrors "github.com/kotalco/api/pkg/errors"
	"regexp"
)

type CreateWorkspaceRequestDto struct {
	Name string `json:"name"  validate:"required,gte=1,lte=100"`
}

type WorkspaceResponseDto struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	K8sNamespace string `json:"k8s_namespace"`
}

//Marshall creates workspace response from workspace model
func (dto *WorkspaceResponseDto) Marshall(model *Workspace) *WorkspaceResponseDto {
	dto.ID = model.ID
	dto.Name = model.Name
	dto.K8sNamespace = model.K8sNamespace
	return dto
}

//Validate validates workspace requests fields
func Validate(dto interface{}) *restErrors.RestErr {
	newValidator := validator.New()

	//create custom validation tag for validate with regexp
	err := newValidator.RegisterValidation("regexp", func(fl validator.FieldLevel) bool {
		field := fl.Field().String()
		if field == "" {
			return false
		}
		regexString := fl.Param()
		regex := regexp.MustCompile(regexString)
		match := regex.MatchString(field)
		return match
	})
	if err != nil {
		return restErrors.NewInternalServerError("something went wrong")
	}

	err = newValidator.Struct(dto)

	if err != nil {
		fields := map[string]string{}
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Field() {
			case "Name":
				fields["name"] = "name should be greater than 1 char and less than 100 char"
				break
			}
		}
		if len(fields) > 0 {
			return restErrors.NewValidationError(fields)
		}
	}
	return nil
}
