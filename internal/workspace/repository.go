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
