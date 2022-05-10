package workspace

import (
	"github.com/go-playground/validator/v10"
	restErrors "github.com/kotalco/api/pkg/errors"
)

type CreateWorkspaceRequestDto struct {
	Name string `json:"name" validate:"regexp=[a-z0-9]([-a-z0-9]*[a-z0-9])?"`
}

type WorkspaceResponseDto struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	K8sNamespace string `json:"k8s_Namespace"`
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
	err := newValidator.Struct(dto)

	if err != nil {
		fields := map[string]string{}
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Field() {
			case "Name":
				fields["name"] = "name can only have lowercase, alphanumeric or hyphen values"
				break
			}
		}
		if len(fields) > 0 {
			return restErrors.NewValidationError(fields)
		}
	}
	return nil
}
