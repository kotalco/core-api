package workspace

import (
	"github.com/google/uuid"
	"github.com/kotalco/cloud-api/internal/workspaceuser"
	"github.com/kotalco/cloud-api/pkg/k8s"
	"github.com/kotalco/cloud-api/pkg/roles"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	"gorm.io/gorm"
	"net/http"
)

type service struct {
	db *gorm.DB
}

type IService interface {
	WithTransaction(txHandle *gorm.DB) IService
	WithoutTransaction() IService
	Create(dto *CreateWorkspaceRequestDto, userId string) (*Workspace, restErrors.IRestErr)
	Update(dto *UpdateWorkspaceRequestDto, workspace *Workspace) restErrors.IRestErr
	Delete(workspace *Workspace) restErrors.IRestErr
	GetById(Id string) (*Workspace, restErrors.IRestErr)
	GetByUserId(UserId string) ([]*Workspace, restErrors.IRestErr)
	AddWorkspaceMember(workspace *Workspace, memberId string, role string) restErrors.IRestErr
	DeleteWorkspaceMember(workspace *Workspace, memberId string) restErrors.IRestErr
	CountByUserId(userId string) (int64, restErrors.IRestErr)
	UpdateWorkspaceUser(workspaceUser *workspaceuser.WorkspaceUser, dto *UpdateWorkspaceUserRequestDto) restErrors.IRestErr
	CreateUserDefaultWorkspace(userId string) restErrors.IRestErr
}

var (
	workspaceRepo    = NewRepository()
	namespaceService = k8s.NewNamespaceService()
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
func (service) Create(dto *CreateWorkspaceRequestDto, userId string) (*Workspace, restErrors.IRestErr) {
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

// CreateUserDefaultWorkspace creates a default workspace for the user , or err if any
// it finds the default namespace if it doesn't exist , create namespace with the name default
// if the default namespace already bound to workspace, creates another namespace with  randomName
// creates the default workspace
// this should only be used once in the user registration scenario
func (service) CreateUserDefaultWorkspace(userId string) restErrors.IRestErr {
	defaultNamespace := "default"

	//check if user don't have any workspaces
	list, err := workspaceRepo.GetByUserId(userId)
	if err != nil {
		return err
	}
	if len(list) > 0 {
		return restErrors.NewConflictError("user already have a workspace")
	}

	//check if the cluster has default namespace with the name default, create if it doesn't exist
	_, err = namespaceService.Get(defaultNamespace)
	if err != nil {
		if err.StatusCode() == http.StatusNotFound { //cluster don't have default namespace create one
			err = namespaceService.Create(defaultNamespace)
			if err != nil {
				go logger.Error(service.CreateUserDefaultWorkspace, err)
				return restErrors.NewInternalServerError("can't create the namespace default")
			}
		} else {
			return err
		}
	}

	//get the default workspace that should be bound with the namespace default
	workspaceModel, _ := workspaceRepo.GetByNamespace(defaultNamespace)

	if workspaceModel != nil { // there is already a workspace(user) bound to the namespace "default" , create another namespace to be the default namespace for this user
		defaultNamespace = uuid.NewString()
		err = namespaceService.Create(defaultNamespace)
		if err != nil {
			go logger.Error(service.CreateUserDefaultWorkspace, err)
			return restErrors.NewInternalServerError("can't create the user default namespace")
		}
	}

	//create the default workspace
	workspace := new(Workspace)
	workspace.ID = uuid.New().String()
	workspace.Name = "default"
	workspace.K8sNamespace = defaultNamespace
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
		return err
	}

	return nil
}
