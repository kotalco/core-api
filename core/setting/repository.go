package setting

import (
	"fmt"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/logger"
	"github.com/kotalco/core-api/pkg/sqlclient"
	"gorm.io/gorm"
	"regexp"
)

type repository struct {
	db *gorm.DB
}

type IRepository interface {
	WithTransaction(txHandle *gorm.DB) IRepository
	WithoutTransaction() IRepository
	Get(key string) (string, restErrors.IRestErr)
	Create(key string, value string) restErrors.IRestErr
	Update(key string, value string) restErrors.IRestErr
	Find() ([]*Setting, restErrors.IRestErr)
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

func (r repository) Get(key string) (string, restErrors.IRestErr) {
	var record = new(Setting)

	result := r.db.Where("key = ?", key).First(record)
	if result.Error != nil {
		return "", restErrors.NewNotFoundError(fmt.Sprintf("can't find config for the key  %s", key))
	}

	return record.Value, nil
}

func (r repository) Create(key string, value string) restErrors.IRestErr {
	var record = &Setting{
		Key:   key,
		Value: value,
	}

	res := r.db.Create(record)
	if res.Error != nil {
		duplicateKey, _ := regexp.Match("duplicate key", []byte(res.Error.Error()))
		if duplicateKey {
			return restErrors.NewConflictError("key already exists")
		}
		go logger.Error(repository.Create, res.Error)
		return restErrors.NewInternalServerError("can't create config")
	}

	return nil
}

func (r repository) Update(key string, value string) restErrors.IRestErr {
	result := r.db.Model(Setting{}).Where("key = ?", key).Update("value", value)
	if result.Error != nil {
		go logger.Error(r.Update, result.Error)
		return restErrors.NewInternalServerError("something went wrong")
	}
	return nil
}

func (r repository) Find() ([]*Setting, restErrors.IRestErr) {
	var setting []*Setting

	result := r.db.Find(&setting)
	if result.Error != nil {
		return nil, restErrors.NewNotFoundError(fmt.Sprintf("can't get settings"))
	}

	return setting, nil
}
