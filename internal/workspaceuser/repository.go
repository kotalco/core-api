package workspaceuser

import (
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	"gorm.io/gorm"
)

type repository struct{}

type IRepository interface {
	WithTransaction(txHandle *gorm.DB) IRepository
	Create(workspace *WorkspaceUser) *restErrors.RestErr
}

func NewRepository() IRepository {
	return &repository{}
}

func (repo *repository) WithTransaction(txHandle *gorm.DB) IRepository {
	sqlclient.DbClient = txHandle
	return repo
}

func (repo *repository) Create(workspaceuser *WorkspaceUser) *restErrors.RestErr {
	res := sqlclient.DbClient.Create(workspaceuser)
	if res.Error != nil {
		go logger.Error(repo.Create, res.Error)
		return restErrors.NewInternalServerError("can't create workspace user record")
	}

	return nil
}
