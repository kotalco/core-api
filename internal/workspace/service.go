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
	Update(dto *UpdateWorkspaceRequestDto, workspace *Workspace) *restErrors.RestErr
	Delete(workspace *Workspace) *restErrors.RestErr
	GetById(Id string) (*Workspace, *restErrors.RestErr)
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

//Create creates new workspace and workspace-user record  from a given dto.
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

//Update updates the workspace name only ,updating K8sNamespace not allowed.
func (service) Update(dto *UpdateWorkspaceRequestDto, workspace *Workspace) *restErrors.RestErr {
	workspace.Name = dto.Name
	err := workspaceRepo.Update(workspace)
	if err != nil {
		return err
	}

	return nil
}

//GetById gets workspace by given id.
func (service) GetById(Id string) (*Workspace, *restErrors.RestErr) {
	return workspaceRepo.GetById(Id)
}

//Delete workspace by given id for specific user
func (service) Delete(workspace *Workspace) *restErrors.RestErr {
	err := workspaceRepo.Delete(workspace)
	if err != nil {
		return err
	}

	return nil
}
