package endpointactivity

import (
	"errors"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	"gorm.io/gorm"
)

var (
	queryGetByEndpointId = "endpoint_id = ?"
)

type repository struct {
	db *gorm.DB
}

type IRepository interface {
	WithTransaction(txHandle *gorm.DB) IRepository
	WithoutTransaction() IRepository
	Update(activity *Activity) restErrors.IRestErr
	FindOne(query interface{}, conditions ...interface{}) (*Activity, restErrors.IRestErr)
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

func (r repository) Update(activity *Activity) restErrors.IRestErr {
	res := r.db.Save(activity)
	if res.Error != nil {
		go logger.Error(r.Update, res.Error)
		return restErrors.NewInternalServerError("something went wrong")
	}
	return nil
}

func (r repository) FindOne(query interface{}, conditions ...interface{}) (*Activity, restErrors.IRestErr) {
	var record = new(Activity)

	result := r.db.Where(query, conditions).First(record)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, restErrors.NewNotFoundError("record not found")
		}
		go logger.Error(r.FindOne, result.Error)
		return nil, restErrors.NewInternalServerError("something went wrong")
	}

	return record, nil
}
