package endpointactivity

import (
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	"gorm.io/gorm"
)

var (
	queryGetMonthlyActivity    = "date_trunc('month', (timestamp 'epoch' + timestamp * interval '1 second')) = date_trunc('month', current_date) AND endpoint_id = ?"
	queryGetUserMinuteActivity = "date_trunc('minute', to_timestamp(timestamp)) = date_trunc('minute', current_timestamp) AND user_id = ?"
)

type repository struct {
	db *gorm.DB
}

type IRepository interface {
	WithTransaction(txHandle *gorm.DB) IRepository
	WithoutTransaction() IRepository
	CreateInBatches(activities []*Activity) restErrors.IRestErr
	FindMany(query interface{}, conditions ...interface{}) ([]*Activity, restErrors.IRestErr)
	Count(query interface{}, conditions ...interface{}) (int64, restErrors.IRestErr)
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

func (r repository) CreateInBatches(activities []*Activity) restErrors.IRestErr {
	res := r.db.CreateInBatches(activities, 10)
	if res.Error != nil {
		go logger.Error(r.CreateInBatches, res.Error)
		return restErrors.NewInternalServerError("something went wrong")
	}
	return nil
}

func (r repository) FindMany(query interface{}, conditions ...interface{}) ([]*Activity, restErrors.IRestErr) {
	var records []*Activity
	result := r.db.Where(query, conditions...).Find(&records)
	if result.Error != nil {
		go logger.Error(repository.FindMany, result.Error)
		return nil, restErrors.NewInternalServerError("something went wrong")
	}
	return records, nil
}

func (r repository) Count(query interface{}, conditions ...interface{}) (int64, restErrors.IRestErr) {
	var count int64
	result := r.db.Table("activities").Where(query, conditions...).Count(&count)
	if result.Error != nil {
		go logger.Error(repository.Count, result.Error)
		return 0, restErrors.NewInternalServerError("something went wrong")
	}
	return count, nil
}
