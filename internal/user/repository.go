package user

import (
	"fmt"
	"net/http"
	"regexp"

	"gorm.io/gorm"

	"github.com/kotalco/cloud-api/pkg/sqlclient"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
)

type repository struct{}

var sqlClient = sqlclient.DbClient

type IRepository interface {
	WithTransaction(txHandle gorm.DB) IRepository
	Create(user *User) *restErrors.RestErr
	GetByEmail(email string) (*User, *restErrors.RestErr)
	GetById(id string) (*User, *restErrors.RestErr)
	Update(user *User) *restErrors.RestErr
	FindWhereIdInSlice(ids []string) ([]*User, *restErrors.RestErr)
	Count() (int64, *restErrors.RestErr)
}

func NewRepository() IRepository {
	newRepo := repository{}
	return newRepo
}

func (r repository) WithTransaction(txHandle gorm.DB) IRepository {
	sqlClient = txHandle
	return r
}

func (repository) Create(user *User) *restErrors.RestErr {
	res := sqlClient.Create(user)
	if res.Error != nil {
		duplicateEmail, _ := regexp.Match("duplicate key", []byte(res.Error.Error()))
		if duplicateEmail {
			//todo create conflict error in error pkg
			return &restErrors.RestErr{
				Message: "email already exits",
				Status:  http.StatusConflict,
				Name:    "Conflict",
			}
		}
		go logger.Error(repository.Create, res.Error)
		return restErrors.NewInternalServerError("can't create user")
	}

	return nil
}

func (repository) GetByEmail(email string) (*User, *restErrors.RestErr) {
	var user = new(User)

	result := sqlClient.Where("email = ?", email).First(user)
	if result.Error != nil {
		return nil, restErrors.NewNotFoundError(fmt.Sprintf("can't find user with email  %s", email))
	}

	return user, nil
}

func (repository) GetById(id string) (*User, *restErrors.RestErr) {
	var user = new(User)

	result := sqlClient.Where("id = ?", id).First(user)
	if result.Error != nil {
		return nil, restErrors.NewNotFoundError(fmt.Sprintf("no such user"))
	}

	return user, nil
}

func (repository) Update(user *User) *restErrors.RestErr {
	err := sqlClient.Save(user)
	if err.Error != nil {
		go logger.Error(repository.Update, err.Error)
		return restErrors.NewInternalServerError("something went wrong")
	}

	return nil
}

func (repository) FindWhereIdInSlice(ids []string) ([]*User, *restErrors.RestErr) {
	var users []*User
	result := sqlClient.Where("id IN (?)", ids).Find(&users)
	if result.Error != nil {
		go logger.Error(repository.FindWhereIdInSlice, result.Error)
		return nil, restErrors.NewInternalServerError("something went wrong")
	}
	return users, nil
}

func (repository) Count() (int64, *restErrors.RestErr) {
	var count int64
	result := sqlClient.Model(User{}).Count(&count)
	if result.Error != nil {
		go logger.Error(repository.Count, result.Error)
		return 0, restErrors.NewInternalServerError("can't count users")
	}
	return count, nil
}
