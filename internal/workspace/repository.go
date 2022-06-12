package workspace

import (
	"errors"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/cloud-api/internal/workspaceuser"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	"gorm.io/gorm"
)

type repository struct{}

type IRepository interface {
	WithTransaction(txHandle *gorm.DB) IRepository
	Create(workspace *Workspace) *restErrors.RestErr
	GetByNameAndUserId(name string, userId string) ([]*Workspace, *restErrors.RestErr)
	GetById(id string) (*Workspace, *restErrors.RestErr)
	Update(workspace *Workspace) *restErrors.RestErr
	Delete(*Workspace) *restErrors.RestErr
	GetByUserId(userId string) ([]*Workspace, *restErrors.RestErr)
	AddWorkspaceMember(workspace *Workspace, workspaceUser *workspaceuser.WorkspaceUser) *restErrors.RestErr
	DeleteWorkspaceMember(workspace *Workspace, workspaceUser *workspaceuser.WorkspaceUser) *restErrors.RestErr
	GetWorkspaceMemberByWorkspaceIdAndUserId(workspaceId string, userId string) (*workspaceuser.WorkspaceUser, *restErrors.RestErr)
	CountByUserId(userId string) (int64, *restErrors.RestErr)
}

func NewRepository() IRepository {
	return &repository{}
}

func (repo *repository) WithTransaction(txHandle *gorm.DB) IRepository {
	sqlclient.DbClient = txHandle
	return repo
}

//Create creates a new workspace record with its first workspaceUser record
func (repo *repository) Create(workspace *Workspace) *restErrors.RestErr {
	res := sqlclient.DbClient.Create(workspace)
	if res.Error != nil {
		go logger.Error(repo.Create, res.Error)
		return restErrors.NewInternalServerError("can't create workspace")
	}

	return nil
}

//GetByNameAndUserId used to get workspace by name to check if workspace name exits for the same owner(userId)
func (repo *repository) GetByNameAndUserId(name string, userId string) ([]*Workspace, *restErrors.RestErr) {
	var workspaces []*Workspace
	result := sqlclient.DbClient.Where("user_id = ? AND name = ?", userId, name).Find(&workspaces)
	if result.Error != nil {
		go logger.Error(repo.GetByNameAndUserId, result.Error)
		return nil, restErrors.NewInternalServerError("something went wrong")
	}
	return workspaces, nil
}

//Update updates workspace record
func (repo *repository) Update(workspace *Workspace) *restErrors.RestErr {
	res := sqlclient.DbClient.Save(workspace)
	if res.Error != nil {
		go logger.Error(repo.Update, res.Error)
		return restErrors.NewInternalServerError("something went wrong")
	}

	return nil
}

//GetById gets workspace record and preloads workspaceUser records
func (repo repository) GetById(Id string) (*Workspace, *restErrors.RestErr) {
	var workspace = new(Workspace)
	workspace.ID = Id

	result := sqlclient.DbClient.Preload("WorkspaceUsers").First(workspace)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, restErrors.NewNotFoundError("record not found")
		}
		go logger.Error(repo.GetById, result.Error)
		return nil, restErrors.NewInternalServerError("something went wrong")
	}

	return workspace, nil
}

//Delete deletes workspace record and cascades workspaceUser records with it
func (repo repository) Delete(workspace *Workspace) *restErrors.RestErr {
	result := sqlclient.DbClient.First(workspace).Delete(workspace)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return restErrors.NewNotFoundError("record not found")
		}
		go logger.Error(repo.Delete, result.Error)
		return restErrors.NewInternalServerError("something went wrong")
	}

	return nil
}

//GetByUserId get workspaces where user assigned to member or owner by sub-query over workspaceUser table
func (repo repository) GetByUserId(userId string) ([]*Workspace, *restErrors.RestErr) {
	var workspaces []*Workspace
	subQuery := sqlclient.DbClient.Model(workspaceuser.WorkspaceUser{}).Where("user_id = ?", userId).Select("workspace_id")
	result := sqlclient.DbClient.Where("id IN (?)", subQuery).Find(&workspaces)
	if result.Error != nil {
		go logger.Error(repo.GetByUserId, result.Error)
		return nil, restErrors.NewInternalServerError("something went wrong")
	}

	return workspaces, nil
}

//AddWorkspaceMember create workspaceUser record through association with workspace
func (repo *repository) AddWorkspaceMember(workspace *Workspace, workspaceUser *workspaceuser.WorkspaceUser) *restErrors.RestErr {
	err := sqlclient.DbClient.Model(workspace).Association("WorkspaceUsers").Append(workspaceUser)
	if err != nil {
		go logger.Error(repo.AddWorkspaceMember, err)
		return restErrors.NewInternalServerError("something went wrong")
	}

	return nil
}

//DeleteWorkspaceMember removes existing workspaceUser record through association with workspace
func (repo *repository) DeleteWorkspaceMember(workspace *Workspace, workspaceUser *workspaceuser.WorkspaceUser) *restErrors.RestErr {
	err := sqlclient.DbClient.Model(workspace).Association("WorkspaceUsers").Delete(workspaceUser)
	if err != nil {
		go logger.Error(repo.DeleteWorkspaceMember, err)
		return restErrors.NewInternalServerError("something went wrong")
	}
	result := sqlclient.DbClient.Delete(workspaceUser)
	if result.Error != nil {
		go logger.Error(repo.DeleteWorkspaceMember, result.Error)
		return restErrors.NewInternalServerError("something went wrong")
	}

	return nil
}

//GetWorkspaceMemberByWorkspaceIdAndUserId finds workspace member by workspaceId and userId
func (repo *repository) GetWorkspaceMemberByWorkspaceIdAndUserId(workspaceId string, userId string) (*workspaceuser.WorkspaceUser, *restErrors.RestErr) {
	var workspaceUser = new(workspaceuser.WorkspaceUser)
	result := sqlclient.DbClient.Where("user_id = ? AND workspace_id = ?", userId, workspaceId).First(workspaceUser)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, restErrors.NewNotFoundError("record not found")
		}
		go logger.Error(repo.GetWorkspaceMemberByWorkspaceIdAndUserId, result.Error)
		return nil, restErrors.NewInternalServerError("something went wrong")
	}
	return workspaceUser, nil
}

//CountByUserId returns user's workspaces count
func (repo *repository) CountByUserId(userId string) (int64, *restErrors.RestErr) {
	var count int64
	result := sqlclient.DbClient.Model(Workspace{}).Where("user_id = ?", userId).Count(&count)
	if result.Error != nil {
		go logger.Error(repo.CountByUserId, result.Error)
		return 0, restErrors.NewInternalServerError("something went wrong")
	}
	return count, nil
}
