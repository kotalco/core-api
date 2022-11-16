package dbconfig

import (
	"fmt"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	"gorm.io/gorm"
	"net/http"
	"regexp"
)

type repository struct{}

type IRepository interface {
	WithTransaction(txHandle *gorm.DB) IRepository
	Get(key string) (string, *restErrors.RestErr)
	Set(key string, value string) *restErrors.RestErr
}

func NewRepository() IRepository {
	newRepo := repository{}
	return newRepo
}

func (r repository) WithTransaction(txHandle *gorm.DB) IRepository {
	sqlclient.DbClient = txHandle
	return r
}

func (r repository) Get(key string) (string, *restErrors.RestErr) {
	var record = new(DbConfig)

	result := sqlclient.DbClient.Where("key = ?", key).First(record)
	if result.Error != nil {
		return "", restErrors.NewNotFoundError(fmt.Sprintf("can't find config for the key  %s", key))
	}

	return key, nil
}

func (r repository) Set(key string, value string) *restErrors.RestErr {
	var record = &DbConfig{
		Key:   key,
		Value: value,
	}

	res := sqlclient.DbClient.Create(record)
	if res.Error != nil {
		duplicateEmail, _ := regexp.Match("duplicate key", []byte(res.Error.Error()))
		if duplicateEmail {
			//todo create conflict error in error pkg
			return &restErrors.RestErr{
				Message: "key already exists",
				Status:  http.StatusConflict,
				Name:    "Conflict",
			}
		}
		go logger.Error(repository.Set, res.Error)
		return restErrors.NewInternalServerError("can't create config")
	}

	return nil
}
