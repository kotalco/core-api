package workspace

import (
	"github.com/google/uuid"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/cloud-api/internal/workspaceuser"
	"gorm.io/gorm"
	"net/http"
)

type service struct{}

type IService interface {
	Create(dto *CreateWorkspaceRequestDto, userId string) (*WorkspaceResponseDto, *restErrors.RestErr)
	WithTransaction(txHandle *gorm.DB) IService
}

var (
	workspaceRepo     = NewRepository()
	workspaceUserRepo = workspaceuser.NewRepository()
)

func NewService() IService {
	return &service{}
}

func (wService service) WithTransaction(txHandle *gorm.DB) IService {
	workspaceRepo = workspaceRepo.WithTransaction(txHandle)
	return wService
}

func (service) Create(dto *CreateWorkspaceRequestDto, userId string) (*WorkspaceResponseDto, *restErrors.RestErr) {
	if dto.Name == "" {
		dto.Name = "default"
	}

	exits, err := workspaceRepo.GetByNameAndUserId(dto.Name, userId)
	if err != nil && err.Status != http.StatusNotFound {
		return nil, err
	}
	if exits != nil {
		return nil, restErrors.NewConflictError("workspace already exits")
	}

	workspace := new(Workspace)
	workspace.ID = uuid.New().String()

	workspace.Name = dto.Name
	workspace.K8sNamespace = uuid.New().String()
	workspace.UserId = userId

	err = workspaceRepo.Create(workspace)
	if err != nil {
		return nil, err
	}

	workspaceuser := new(workspaceuser.WorkspaceUser)
	workspaceuser.ID = uuid.New().String()
	workspaceuser.WorkspaceId = workspace.ID
	workspaceuser.UserId = workspace.UserId

	err = workspaceUserRepo.Create(workspaceuser)
	if err != nil {
		return nil, err
	}

	return new(WorkspaceResponseDto).Marshall(workspace), nil
}
