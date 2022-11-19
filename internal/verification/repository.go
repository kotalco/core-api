package verification

import (
	"fmt"
	"gorm.io/gorm"
	"regexp"

	"github.com/kotalco/cloud-api/pkg/sqlclient"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
)

type repository struct{}

type IRepository interface {
	WithTransaction(txHandle *gorm.DB) IRepository
	Create(verification *Verification) *restErrors.RestErr
	GetByUserId(userId string) (*Verification, *restErrors.RestErr)
	Update(verification *Verification) *restErrors.RestErr
}

func NewRepository() IRepository {
	sqlclient.OpenDBConnection()
	newRepository := repository{}
	return newRepository
}

func (uRepository repository) WithTransaction(txHandle *gorm.DB) IRepository {
	sqlclient.DbClient = txHandle
	return uRepository
}

func (repository) Create(verification *Verification) *restErrors.RestErr {
	res := sqlclient.DbClient.Create(verification)
	if res.Error != nil {
		duplicateEmail, _ := regexp.Match("duplicate key", []byte(res.Error.Error()))
		if duplicateEmail {
			return restErrors.NewBadRequestError("verification already exits")
		}
		go logger.Error(repository.Create, res.Error)
		return restErrors.NewInternalServerError("can't create verification")
	}

	return nil
}

func (repository) GetByUserId(userId string) (*Verification, *restErrors.RestErr) {
	var verification = new(Verification)

	result := sqlclient.DbClient.Where("user_id = ?", userId).First(verification)
	if result.Error != nil {
		return nil, restErrors.NewNotFoundError(fmt.Sprintf("can't find verification with userId  %s", userId))
	}

	return verification, nil
}

func (repository) Update(verification *Verification) *restErrors.RestErr {
	resp := sqlclient.DbClient.Save(verification)
	if resp.Error != nil {
		go logger.Error(repository.Update, resp.Error)
		return restErrors.NewInternalServerError("some thing went wrong!")
	}

	return nil
}
