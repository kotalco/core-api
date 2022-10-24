package workspace

import (
	"github.com/google/uuid"
	"github.com/kotalco/cloud-api/internal/workspaceuser"
	"github.com/kotalco/cloud-api/pkg/roles"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"gorm.io/gorm"
	"net/http"
)

type service struct{}

type IService interface {
	WithTransaction(txHandle *gorm.DB) IService
	Create(dto *CreateWorkspaceRequestDto, userId string) (*Workspace, *restErrors.RestErr)
	Update(dto *UpdateWorkspaceRequestDto, workspace *Workspace) *restErrors.RestErr
	Delete(workspace *Workspace) *restErrors.RestErr
	GetById(Id string) (*Workspace, *restErrors.RestErr)
	GetByUserId(UserId string) ([]*Workspace, *restErrors.RestErr)
	AddWorkspaceMember(workspace *Workspace, memberId string, role string) *restErrors.RestErr
	DeleteWorkspaceMember(workspace *Workspace, memberId string) *restErrors.RestErr
	CountByUserId(userId string) (int64, *restErrors.RestErr)
	UpdateWorkspaceUser(workspaceUser *workspaceuser.WorkspaceUser, dto *UpdateWorkspaceUserRequestDto) *restErrors.RestErr
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

// Create creates new workspace and workspace-user record  from a given dto.
func (service) Create(dto *CreateWorkspaceRequestDto, userId string) (*Workspace, *restErrors.RestErr) {
	exist, err := workspaceRepo.GetByNameAndUserId(dto.Name, userId)
	if err != nil && err.Status != http.StatusNotFound {
		return nil, err
	}

	if len(exist) > 0 {
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
	workspaceuser.Role = roles.Admin

	workspace.WorkspaceUsers = append(workspace.WorkspaceUsers, *workspaceuser)

	err = workspaceRepo.Create(workspace)
	if err != nil {
		return nil, err
	}

	return workspace, nil
}

// Update updates the workspace name only ,updating K8sNamespace not allowed.
func (service) Update(dto *UpdateWorkspaceRequestDto, workspace *Workspace) *restErrors.RestErr {
	exist, err := workspaceRepo.GetByNameAndUserId(dto.Name, workspace.UserId)
	if err != nil && err.Status != http.StatusNotFound {
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
func (service) GetById(Id string) (*Workspace, *restErrors.RestErr) {
	return workspaceRepo.GetById(Id)
}

// Delete workspace by given id for specific user
func (service) Delete(workspace *Workspace) *restErrors.RestErr {
	err := workspaceRepo.Delete(workspace)
	if err != nil {
		return err
	}

	return nil
}

// GetByUserId Finds workspaces by given userId
func (service) GetByUserId(userId string) ([]*Workspace, *restErrors.RestErr) {
	return workspaceRepo.GetByUserId(userId)
}

// AddWorkspaceMember creates new workspaceUser record for a given workspace
func (service) AddWorkspaceMember(workspace *Workspace, memberId string, role string) *restErrors.RestErr {
	newWorkspaceUser := new(workspaceuser.WorkspaceUser)
	newWorkspaceUser.ID = uuid.NewString()
	newWorkspaceUser.WorkspaceID = workspace.ID
	newWorkspaceUser.UserId = memberId
	newWorkspaceUser.Role = role

	return workspaceRepo.AddWorkspaceMember(workspace, newWorkspaceUser)
}

// DeleteWorkspaceMember removes workspaceUser record for a given workspace
func (service) DeleteWorkspaceMember(workspace *Workspace, memberId string) *restErrors.RestErr {
	member, err := workspaceRepo.GetWorkspaceMemberByWorkspaceIdAndUserId(workspace.ID, memberId)
	if err != nil {
		return err
	}

	return workspaceRepo.DeleteWorkspaceMember(workspace, member)
}

// CountByUserId returns user's workspaces count
func (service) CountByUserId(userId string) (int64, *restErrors.RestErr) {
	return workspaceRepo.CountByUserId(userId)
}

func (service) UpdateWorkspaceUser(workspaceUser *workspaceuser.WorkspaceUser, dto *UpdateWorkspaceUserRequestDto) *restErrors.RestErr {
	workspaceUser.Role = dto.Role
	return workspaceRepo.UpdateWorkspaceUser(workspaceUser)
}
