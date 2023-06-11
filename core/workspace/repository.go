package workspace

import (
	"errors"
	"github.com/kotalco/cloud-api/core/workspaceuser"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

type IRepository interface {
	WithTransaction(txHandle *gorm.DB) IRepository
	WithoutTransaction() IRepository
	Create(workspace *Workspace) restErrors.IRestErr
	GetByNameAndUserId(name string, userId string) ([]*Workspace, restErrors.IRestErr)
	GetById(id string) (*Workspace, restErrors.IRestErr)
	Update(workspace *Workspace) restErrors.IRestErr
	Delete(*Workspace) restErrors.IRestErr
	GetByUserId(userId string) ([]*Workspace, restErrors.IRestErr)
	AddWorkspaceMember(workspace *Workspace, workspaceUser *workspaceuser.WorkspaceUser) restErrors.IRestErr
	DeleteWorkspaceMember(workspace *Workspace, workspaceUser *workspaceuser.WorkspaceUser) restErrors.IRestErr
	GetWorkspaceMemberByWorkspaceIdAndUserId(workspaceId string, userId string) (*workspaceuser.WorkspaceUser, restErrors.IRestErr)
	CountByUserId(userId string) (int64, restErrors.IRestErr)
	UpdateWorkspaceUser(workspaceUser *workspaceuser.WorkspaceUser) restErrors.IRestErr
	GetByNamespace(namespace string) (*Workspace, restErrors.IRestErr)
}

func NewRepository() IRepository {
	return &repository{
		db: sqlclient.OpenDBConnection(),
	}
}

func (repo *repository) WithTransaction(txHandle *gorm.DB) IRepository {
	repo.db = txHandle
	return repo
}
func (repo *repository) WithoutTransaction() IRepository {
	repo.db = sqlclient.OpenDBConnection()
	return repo
}

// Create creates a new workspace record with its first workspaceUser record
func (repo *repository) Create(workspace *Workspace) restErrors.IRestErr {
	res := repo.db.Create(workspace)
	if res.Error != nil {
		go logger.Error(repo.Create, res.Error)
		return restErrors.NewInternalServerError("can't create workspace")
	}

	return nil
}

// GetByNameAndUserId used to get workspace by name to check if workspace name exits for the same owner(userId)
func (repo *repository) GetByNameAndUserId(name string, userId string) ([]*Workspace, restErrors.IRestErr) {
	var workspaces []*Workspace
	result := repo.db.Where("user_id = ? AND name = ?", userId, name).Find(&workspaces)
	if result.Error != nil {
		go logger.Error(repo.GetByNameAndUserId, result.Error)
		return nil, restErrors.NewInternalServerError("something went wrong")
	}
	return workspaces, nil
}

// Update updates workspace record
func (repo *repository) Update(workspace *Workspace) restErrors.IRestErr {
	res := repo.db.Save(workspace)
	if res.Error != nil {
		go logger.Error(repo.Update, res.Error)
		return restErrors.NewInternalServerError("something went wrong")
	}

	return nil
}

// GetById gets workspace record and preloads workspaceUser records
func (repo repository) GetById(Id string) (*Workspace, restErrors.IRestErr) {
	var workspace = new(Workspace)
	workspace.ID = Id

	result := repo.db.Preload("WorkspaceUsers").First(workspace)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, restErrors.NewNotFoundError("record not found")
		}
		go logger.Error(repo.GetById, result.Error)
		return nil, restErrors.NewInternalServerError("something went wrong")
	}

	return workspace, nil
}

// Delete deletes workspace record and cascades workspaceUser records with it
func (repo repository) Delete(workspace *Workspace) restErrors.IRestErr {
	result := repo.db.First(workspace).Delete(workspace)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return restErrors.NewNotFoundError("record not found")
		}
		go logger.Error(repo.Delete, result.Error)
		return restErrors.NewInternalServerError("something went wrong")
	}

	return nil
}

// GetByUserId get workspaces where user assigned to member or owner by sub-query over workspaceUser table
func (repo repository) GetByUserId(userId string) ([]*Workspace, restErrors.IRestErr) {
	var workspaces []*Workspace
	subQuery := repo.db.Model(workspaceuser.WorkspaceUser{}).Where("user_id = ?", userId).Select("workspace_id")
	result := repo.db.Where("id IN (?)", subQuery).Find(&workspaces)
	if result.Error != nil {
		go logger.Error(repo.GetByUserId, result.Error)
		return nil, restErrors.NewInternalServerError("something went wrong")
	}

	return workspaces, nil
}

// AddWorkspaceMember create workspaceUser record through association with workspace
func (repo *repository) AddWorkspaceMember(workspace *Workspace, workspaceUser *workspaceuser.WorkspaceUser) restErrors.IRestErr {
	err := repo.db.Model(workspace).Association("WorkspaceUsers").Append(workspaceUser)
	if err != nil {
		go logger.Error(repo.AddWorkspaceMember, err)
		return restErrors.NewInternalServerError("something went wrong")
	}

	return nil
}

// DeleteWorkspaceMember removes existing workspaceUser record through association with workspace
func (repo *repository) DeleteWorkspaceMember(workspace *Workspace, workspaceUser *workspaceuser.WorkspaceUser) restErrors.IRestErr {
	err := repo.db.Model(workspace).Association("WorkspaceUsers").Delete(workspaceUser)
	if err != nil {
		go logger.Error(repo.DeleteWorkspaceMember, err)
		return restErrors.NewInternalServerError("something went wrong")
	}
	result := repo.db.Delete(workspaceUser)
	if result.Error != nil {
		go logger.Error(repo.DeleteWorkspaceMember, result.Error)
		return restErrors.NewInternalServerError("something went wrong")
	}

	return nil
}

// GetWorkspaceMemberByWorkspaceIdAndUserId finds workspace member by workspaceId and userId
func (repo *repository) GetWorkspaceMemberByWorkspaceIdAndUserId(workspaceId string, userId string) (*workspaceuser.WorkspaceUser, restErrors.IRestErr) {
	var workspaceUser = new(workspaceuser.WorkspaceUser)
	result := repo.db.Where("user_id = ? AND workspace_id = ?", userId, workspaceId).First(workspaceUser)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, restErrors.NewNotFoundError("record not found")
		}
		go logger.Error(repo.GetWorkspaceMemberByWorkspaceIdAndUserId, result.Error)
		return nil, restErrors.NewInternalServerError("something went wrong")
	}
	return workspaceUser, nil
}

// CountByUserId returns user's workspaces count
func (repo *repository) CountByUserId(userId string) (int64, restErrors.IRestErr) {
	var count int64
	result := repo.db.Model(Workspace{}).Where("user_id = ?", userId).Count(&count)
	if result.Error != nil {
		go logger.Error(repo.CountByUserId, result.Error)
		return 0, restErrors.NewInternalServerError("something went wrong")
	}
	return count, nil
}

// UpdateWorkspaceUser updates work space user details
func (repo *repository) UpdateWorkspaceUser(workspaceUser *workspaceuser.WorkspaceUser) restErrors.IRestErr {
	res := repo.db.Save(workspaceUser)
	if res.Error != nil {
		go logger.Error(repo.UpdateWorkspaceUser, res.Error)
		return restErrors.NewInternalServerError("something went wrong")
	}
	return nil
}

// GetByNamespace returns workspace by namespace
func (repo *repository) GetByNamespace(namespace string) (*Workspace, restErrors.IRestErr) {
	var workspace = new(Workspace)
	workspace.K8sNamespace = namespace

	result := repo.db.Preload("WorkspaceUsers").Where("k8s_namespace = ?", namespace).First(workspace)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, restErrors.NewNotFoundError("record not found")
		}
		go logger.Error(repo.GetByNamespace, result.Error)
		return nil, restErrors.NewInternalServerError("something went wrong")
	}

	return workspace, nil
}
