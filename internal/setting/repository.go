package setting

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
	Create(key string, value string) *restErrors.RestErr
	Update(key string, value string) *restErrors.RestErr
	Find() ([]*Setting, *restErrors.RestErr)
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
	var record = new(Setting)

	result := sqlclient.DbClient.Where("key = ?", key).First(record)
	if result.Error != nil {
		return "", restErrors.NewNotFoundError(fmt.Sprintf("can't find config for the key  %s", key))
	}

	return record.Value, nil
}

func (r repository) Create(key string, value string) *restErrors.RestErr {
	var record = &Setting{
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
		go logger.Error(repository.Create, res.Error)
		return restErrors.NewInternalServerError("can't create config")
	}

	return nil
}

func (r repository) Update(key string, value string) *restErrors.RestErr {
	result := sqlclient.DbClient.Model(Setting{}).Where("key = ?", key).Update("value", value)
	if result.Error != nil {
		go logger.Error(r.Update, result.Error)
		return restErrors.NewInternalServerError("something went wrong")
	}
	return nil
}

func (r repository) Find() ([]*Setting, *restErrors.RestErr) {
	var setting []*Setting

	result := sqlclient.DbClient.Find(&setting)
	if result.Error != nil {
		return nil, restErrors.NewNotFoundError(fmt.Sprintf("can't get settings"))
	}

	return setting, nil
}
