package workspace

import (
	"github.com/google/uuid"
	"github.com/kotalco/cloud-api/core/workspaceuser"
	restErrors "github.com/kotalco/cloud-api/pkg/errors"
	"github.com/kotalco/cloud-api/pkg/roles"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	"gorm.io/gorm"
	"net/http"
)

type service struct {
	db *gorm.DB
}

type IService interface {
	WithTransaction(txHandle *gorm.DB) IService
	WithoutTransaction() IService
	Create(dto *CreateWorkspaceRequestDto, userId string, k8sNamespace string) (*Workspace, restErrors.IRestErr)
	Update(dto *UpdateWorkspaceRequestDto, workspace *Workspace) restErrors.IRestErr
	Delete(workspace *Workspace) restErrors.IRestErr
	GetById(Id string) (*Workspace, restErrors.IRestErr)
	GetByUserId(UserId string) ([]*Workspace, restErrors.IRestErr)
	AddWorkspaceMember(workspace *Workspace, memberId string, role string) restErrors.IRestErr
	DeleteWorkspaceMember(workspace *Workspace, memberId string) restErrors.IRestErr
	CountByUserId(userId string) (int64, restErrors.IRestErr)
	UpdateWorkspaceUser(workspaceUser *workspaceuser.WorkspaceUser, dto *UpdateWorkspaceUserRequestDto) restErrors.IRestErr
	GetByNamespace(namespace string) (*Workspace, restErrors.IRestErr)
}

var (
	workspaceRepo = NewRepository()
)

func NewService() IService {
	return &service{
		db: sqlclient.OpenDBConnection(),
	}
}

func (wService service) WithTransaction(txHandle *gorm.DB) IService {
	workspaceRepo = workspaceRepo.WithTransaction(txHandle)
	return wService
}
func (wService service) WithoutTransaction() IService {
	workspaceRepo = workspaceRepo.WithoutTransaction()
	return wService
}

// Create creates new workspace and workspace-user record  from a given dto.
func (service) Create(dto *CreateWorkspaceRequestDto, userId string, k8sNamespace string) (*Workspace, restErrors.IRestErr) {
	exist, err := workspaceRepo.GetByNameAndUserId(dto.Name, userId)
	if err != nil && err.StatusCode() != http.StatusNotFound {
		return nil, err
	}

	if len(exist) > 0 {
		return nil, restErrors.NewConflictError("workspace already exist")
	}

	workspace := new(Workspace)
	workspace.ID = uuid.New().String()
	workspace.Name = dto.Name
	workspace.K8sNamespace = k8sNamespace
	workspace.UserId = userId

	//create workspace-user record
	workspaceuser := new(workspaceuser.WorkspaceUser)
	workspaceuser.ID = uuid.New().String()
	workspaceuser.WorkspaceID = workspace.ID
	workspaceuser.UserId = workspace.UserId
	workspaceuser.Role = roles.Admin

	workspace.WorkspaceUsers = append(workspace.WorkspaceUsers, *workspaceuser)

	err = workspaceRepo.Create(workspace)
	if err != nil {
		return nil, err
	}

	return workspace, nil
}

// Update updates the workspace name only ,updating K8sNamespace not allowed.
func (service) Update(dto *UpdateWorkspaceRequestDto, workspace *Workspace) restErrors.IRestErr {
	exist, err := workspaceRepo.GetByNameAndUserId(dto.Name, workspace.UserId)
	if err != nil && err.StatusCode() != http.StatusNotFound {
		return err
	}

	for _, v := range exist {
		if v.ID != dto.ID {
			return restErrors.NewConflictError("you have another workspace with the same name")

		}
	}

	workspace.Name = dto.Name
	err = workspaceRepo.Update(workspace)
	if err != nil {
		return err
	}

	return nil
}

// GetById gets workspace by given id.
func (service) GetById(Id string) (*Workspace, restErrors.IRestErr) {
	return workspaceRepo.GetById(Id)
}

// Delete workspace by given id for specific user
func (service) Delete(workspace *Workspace) restErrors.IRestErr {
	err := workspaceRepo.Delete(workspace)
	if err != nil {
		return err
	}

	return nil
}

// GetByUserId Finds workspaces by given userId
func (service) GetByUserId(userId string) ([]*Workspace, restErrors.IRestErr) {
	return workspaceRepo.GetByUserId(userId)
}

// AddWorkspaceMember creates new workspaceUser record for a given workspace
func (service) AddWorkspaceMember(workspace *Workspace, memberId string, role string) restErrors.IRestErr {
	newWorkspaceUser := new(workspaceuser.WorkspaceUser)
	newWorkspaceUser.ID = uuid.NewString()
	newWorkspaceUser.WorkspaceID = workspace.ID
	newWorkspaceUser.UserId = memberId
	newWorkspaceUser.Role = role

	return workspaceRepo.AddWorkspaceMember(workspace, newWorkspaceUser)
}

// DeleteWorkspaceMember removes workspaceUser record for a given workspace
func (service) DeleteWorkspaceMember(workspace *Workspace, memberId string) restErrors.IRestErr {
	member, err := workspaceRepo.GetWorkspaceMemberByWorkspaceIdAndUserId(workspace.ID, memberId)
	if err != nil {
		return err
	}

	return workspaceRepo.DeleteWorkspaceMember(workspace, member)
}

// CountByUserId returns user's workspaces count
func (service) CountByUserId(userId string) (int64, restErrors.IRestErr) {
	return workspaceRepo.CountByUserId(userId)
}

func (service) UpdateWorkspaceUser(workspaceUser *workspaceuser.WorkspaceUser, dto *UpdateWorkspaceUserRequestDto) restErrors.IRestErr {
	workspaceUser.Role = dto.Role
	return workspaceRepo.UpdateWorkspaceUser(workspaceUser)
}

func (service) GetByNamespace(namespace string) (*Workspace, restErrors.IRestErr) {
	return workspaceRepo.GetByNamespace(namespace)
}
