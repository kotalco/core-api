package verification

import (
	"fmt"
	"gorm.io/gorm"
	"regexp"

	"github.com/kotalco/cloud-api/pkg/sqlclient"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
)

type repository struct {
	db *gorm.DB
}

type IRepository interface {
	WithTransaction(txHandle *gorm.DB) IRepository
	WithoutTransaction() IRepository
	Create(verification *Verification) *restErrors.RestErr
	GetByUserId(userId string) (*Verification, *restErrors.RestErr)
	Update(verification *Verification) *restErrors.RestErr
}

func NewRepository() IRepository {
	newRepository := repository{}
	return newRepository
}

func (r repository) WithTransaction(txHandle *gorm.DB) IRepository {
	r.db = txHandle
	return r
}
func (r repository) WithoutTransaction() IRepository {
	r.db = sqlclient.OpenDBConnection()
	return r
}

func (r repository) Create(verification *Verification) *restErrors.RestErr {
	res := r.db.Create(verification)
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

func (r repository) GetByUserId(userId string) (*Verification, *restErrors.RestErr) {
	var verification = new(Verification)

	result := r.db.Where("user_id = ?", userId).First(verification)
	if result.Error != nil {
		return nil, restErrors.NewNotFoundError(fmt.Sprintf("can't find verification with userId  %s", userId))
	}

	return verification, nil
}

func (r repository) Update(verification *Verification) *restErrors.RestErr {
	resp := r.db.Save(verification)
	if resp.Error != nil {
		go logger.Error(repository.Update, resp.Error)
		return restErrors.NewInternalServerError("some thing went wrong!")
	}

	return nil
}
