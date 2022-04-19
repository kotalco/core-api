package user

import (
	"fmt"
	"gorm.io/gorm"
	"net/http"
	"regexp"

	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
)

type repository struct{}

type IRepository interface {
	Create(user *User) *restErrors.RestErr
	GetByEmail(email string) (*User, *restErrors.RestErr)
	GetById(id string) (*User, *restErrors.RestErr)
	Update(user *User) *restErrors.RestErr
	WithTransaction(txHandle *gorm.DB) IRepository
}

var (
	dbClient *gorm.DB
)

func NewRepository() IRepository {
	dbClient = sqlclient.OpenDBConnection()
	newRepo := repository{}
	return newRepo
}

func (r repository) WithTransaction(txHandle *gorm.DB) IRepository {
	dbClient = txHandle
	return r
}

func (repository) Create(user *User) *restErrors.RestErr {
	res := dbClient.Create(user)
	if res.Error != nil {
		duplicateEmail, _ := regexp.Match("duplicate key", []byte(res.Error.Error()))
		if duplicateEmail {
			//todo create conflict error in error pkg
			return &restErrors.RestErr{
				Message: "email already exits",
				Status:  http.StatusConflict,
				Error:   "Conflict",
			}
		}
		go logger.Error(repository.Create, res.Error)
		return restErrors.NewInternalServerError("can't create user")
	}

	return nil
}

func (repository) GetByEmail(email string) (*User, *restErrors.RestErr) {
	var user = new(User)

	result := dbClient.Where("email = ?", email).First(user)
	if result.Error != nil {
		return nil, restErrors.NewNotFoundError(fmt.Sprintf("can't find user with email  %s", email))
	}

	return user, nil
}

func (repository) GetById(id string) (*User, *restErrors.RestErr) {
	var user = new(User)

	result := dbClient.Where("id = ?", id).First(user)
	if result.Error != nil {
		return nil, restErrors.NewNotFoundError(fmt.Sprintf("no such user"))
	}

	return user, nil
}

func (repository) Update(user *User) *restErrors.RestErr {
	err := dbClient.Save(user)
	if err.Error != nil {
		go logger.Error(repository.Update, err.Error)
		return restErrors.NewInternalServerError("something went wrong")
	}

	return nil
}
