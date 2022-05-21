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
	Create(dto *CreateWorkspaceRequestDto, userId string) (*Workspace, *restErrors.RestErr)
	Update(dto *UpdateWorkspaceRequestDto, userId string) (*Workspace, *restErrors.RestErr)
	WithTransaction(txHandle *gorm.DB) IService
}

var (
	workspaceRepo = NewRepository()
)

func NewService() IService {
	return &service{}
}

func (wService service) WithTransaction(txHandle *gorm.DB) IService {
	workspaceRepo = workspaceRepo.WithTransaction(txHandle)
	return wService
}

//Create creates new workspace and workspace-user record  from a given dto ,
func (service) Create(dto *CreateWorkspaceRequestDto, userId string) (*Workspace, *restErrors.RestErr) {
	exist, err := workspaceRepo.GetByNameAndUserId(dto.Name, userId)
	if err != nil && err.Status != http.StatusNotFound {
		return nil, err
	}

	if exist != nil {
		return nil, restErrors.NewConflictError("workspace already exist")
	}

	workspace := new(Workspace)
	workspace.ID = uuid.New().String()
	workspace.Name = dto.Name
	workspace.K8sNamespace = uuid.New().String()
	workspace.UserId = userId

	//create workspace-user record
	workspaceuser := new(workspaceuser.WorkspaceUser)
	workspaceuser.ID = uuid.New().String()
	workspaceuser.WorkspaceID = workspace.ID
	workspaceuser.UserId = workspace.UserId

	workspace.WorkspaceUsers = append(workspace.WorkspaceUsers, *workspaceuser)

	err = workspaceRepo.Create(workspace)
	if err != nil {
		return nil, err
	}

	return workspace, nil
}

//Update updates the workspace name only , not allowed updating K8sNamespace ,
func (service) Update(dto *UpdateWorkspaceRequestDto, userId string) (*Workspace, *restErrors.RestErr) {
	workspace, err := workspaceRepo.GetById(dto.ID)
	if err != nil {
		return nil, err
	}
	if workspace.UserId != userId {
		return nil, restErrors.NewNotFoundError("no such record")
	}

	workspace.Name = dto.Name
	err = workspaceRepo.Update(workspace)
	if err != nil {
		return nil, err
	}

	return workspace, nil
}
