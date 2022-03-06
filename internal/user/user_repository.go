package user

import (
	"fmt"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	"gorm.io/gorm"
	"net/http"
	"regexp"
)

type userRepository struct{}

type userRepositoryInterface interface {
	Create(user *User) *restErrors.RestErr
	GetByEmail(email string) (*User, *restErrors.RestErr)
	GetById(id string) (*User, *restErrors.RestErr)
	Update(user *User) *restErrors.RestErr
}

var (
	UserRepository userRepositoryInterface
	dbClient       *gorm.DB
)

func init() {
	UserRepository = &userRepository{}
	dbClient, _ = sqlclient.OpenDBConnection()
}

func (repo userRepository) Create(user *User) *restErrors.RestErr {
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
		go logger.Error(repo.Create, res.Error)
		return restErrors.NewInternalServerError("can't create user")
	}

	return nil
}

func (repo userRepository) GetByEmail(email string) (*User, *restErrors.RestErr) {
	var user = new(User)

	result := dbClient.Where("email = ?", email).First(user)
	if result.Error != nil {
		return nil, restErrors.NewNotFoundError(fmt.Sprintf("can't find user with email  %s", email))
	}

	return user, nil
}

func (repo userRepository) GetById(id string) (*User, *restErrors.RestErr) {
	var user = new(User)

	result := dbClient.Where("id = ?", id).First(user)
	if result.Error != nil {
		return nil, restErrors.NewNotFoundError(fmt.Sprintf("no such user"))
	}

	return user, nil
}

func (repo userRepository) Update(user *User) *restErrors.RestErr {
	err := dbClient.Save(user)
	if err.Error != nil {
		go logger.Error(repo.Update, err.Error)
		return restErrors.NewInternalServerError("something went wrong")
	}

	return nil
}
