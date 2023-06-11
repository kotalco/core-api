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

type repository struct {
	db *gorm.DB
}

type IRepository interface {
	WithTransaction(txHandle *gorm.DB) IRepository
	WithoutTransaction() IRepository
	Create(user *User) restErrors.IRestErr
	GetByEmail(email string) (*User, restErrors.IRestErr)
	GetById(id string) (*User, restErrors.IRestErr)
	Update(user *User) restErrors.IRestErr
	FindWhereIdInSlice(ids []string) ([]*User, restErrors.IRestErr)
	Count() (int64, restErrors.IRestErr)
}

func NewRepository() IRepository {
	newRepo := repository{}
	newRepo.db = sqlclient.OpenDBConnection()
	return newRepo
}

func (r repository) WithTransaction(txHandle *gorm.DB) IRepository {
	r.db = txHandle
	return r
}
func (r repository) WithoutTransaction() IRepository {
	r.db = sqlclient.OpenDBConnection()
	return r
}

func (r repository) Create(user *User) restErrors.IRestErr {
	res := r.db.Create(user)
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

func (r repository) GetByEmail(email string) (*User, restErrors.IRestErr) {
	var user = new(User)

	result := r.db.Where("email = ?", email).First(user)
	if result.Error != nil {
		return nil, restErrors.NewNotFoundError(fmt.Sprintf("can't find user with email  %s", email))
	}

	return user, nil
}

func (r repository) GetById(id string) (*User, restErrors.IRestErr) {
	var user = new(User)

	result := r.db.Where("id = ?", id).First(user)
	if result.Error != nil {
		return nil, restErrors.NewNotFoundError(fmt.Sprintf("no such user"))
	}

	return user, nil
}

func (r repository) Update(user *User) restErrors.IRestErr {
	err := r.db.Save(user)
	if err.Error != nil {
		go logger.Error(repository.Update, err.Error)
		return restErrors.NewInternalServerError("something went wrong")
	}

	return nil
}

func (r repository) FindWhereIdInSlice(ids []string) ([]*User, restErrors.IRestErr) {
	var users []*User
	result := r.db.Where("id IN (?)", ids).Find(&users)
	if result.Error != nil {
		go logger.Error(repository.FindWhereIdInSlice, result.Error)
		return nil, restErrors.NewInternalServerError("something went wrong")
	}
	return users, nil
}

func (r repository) Count() (int64, restErrors.IRestErr) {
	var count int64
	result := r.db.Model(User{}).Count(&count)
	if result.Error != nil {
		go logger.Error(repository.Count, result.Error)
		return 0, restErrors.NewInternalServerError("can't count users")
	}
	return count, nil
}
