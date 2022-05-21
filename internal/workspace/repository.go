package workspace

import (
	"errors"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	"gorm.io/gorm"
)

type repository struct{}

type IRepository interface {
	WithTransaction(txHandle *gorm.DB) IRepository
	Create(workspace *Workspace) *restErrors.RestErr
	GetByNameAndUserId(name string, userId string) (*Workspace, *restErrors.RestErr)
	GetById(id string) (*Workspace, *restErrors.RestErr)
	Update(workspace *Workspace) *restErrors.RestErr
}

func NewRepository() IRepository {
	return &repository{}
}

func (repo *repository) WithTransaction(txHandle *gorm.DB) IRepository {
	sqlclient.DbClient = txHandle
	return repo
}

func (repo *repository) Create(workspace *Workspace) *restErrors.RestErr {
	res := sqlclient.DbClient.Create(workspace)
	if res.Error != nil {
		go logger.Error(repo.Create, res.Error)
		return restErrors.NewInternalServerError("can't create workspace")
	}

	return nil
}

//GetByNameAndUserId used to get workspace by name to check if workspace name exits for the same owner(userId)
func (repo *repository) GetByNameAndUserId(name string, userId string) (*Workspace, *restErrors.RestErr) {
	var workspace = new(Workspace)
	result := sqlclient.DbClient.Where("user_id = ? AND name = ?", userId, name).First(workspace)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, restErrors.NewNotFoundError("record not found")
		}
		go logger.Error(repo.GetByNameAndUserId, result.Error)
		return nil, restErrors.NewInternalServerError("something went wrong")
	}
	return workspace, nil
}

func (repo *repository) Update(workspace *Workspace) *restErrors.RestErr {
	res := sqlclient.DbClient.Save(workspace)
	if res.Error != nil {
		go logger.Error(repo.Update, res.Error)
		return restErrors.NewInternalServerError(res.Error.Error())
	}

	return nil
}

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
