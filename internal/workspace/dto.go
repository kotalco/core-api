package workspace

import (
	"github.com/go-playground/validator/v10"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/cloud-api/pkg/roles"
)

type CreateWorkspaceRequestDto struct {
	Name string `json:"name"  validate:"required,gte=1,lte=100"`
}

type UpdateWorkspaceRequestDto struct {
	ID   string
	Name string `json:"name"  validate:"required,gte=1,lte=100"`
}

type UpdateWorkspaceUserRequestDto struct {
	Role string `json:"role" validate:"roles"`
}

type WorkspaceResponseDto struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	K8sNamespace string `json:"k8s_namespace"`
	Role         string `json:"role,omitempty"`
}

type AddWorkspaceMemberDto struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"roles"`
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
	err := newValidator.RegisterValidation("roles", func(fl validator.FieldLevel) bool {
		return roles.New().Exist(fl.Field().String())
	})
	if err != nil {
		logger.Panic("USER_DTO_VALIDATE", err)
		return restErrors.NewInternalServerError("something went wrong!")
	}
	err = newValidator.Struct(dto)

	if err != nil {
		fields := map[string]string{}
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Field() {
			case "Name":
				fields["name"] = "name should be greater than 1 char and less than 100 char"
				break
			case "Email":
				fields["email"] = "email should be a valid email address"
				break
			case "Role":
				fields["role"] = "invalid role"
				break
			}
		}
		if len(fields) > 0 {
			return restErrors.NewValidationError(fields)
		}
	}
	return nil
}
