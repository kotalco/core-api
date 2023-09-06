package endpointactivity

import (
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	"gorm.io/gorm"
)

var (
	queryGetMonthlyActivity = "date_trunc('month', (timestamp 'epoch' + timestamp * interval '1 second')) = date_trunc('month', current_date) AND endpoint_id = ?"
)

type repository struct {
	db *gorm.DB
}

type IRepository interface {
	WithTransaction(txHandle *gorm.DB) IRepository
	WithoutTransaction() IRepository
	Create(activity *Activity) restErrors.IRestErr
	FindMany(query interface{}, conditions ...interface{}) ([]*Activity, restErrors.IRestErr)
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

func (r repository) Create(activity *Activity) restErrors.IRestErr {
	res := r.db.Create(activity)
	if res.Error != nil {
		go logger.Error(r.Create, res.Error)
		return restErrors.NewInternalServerError("something went wrong")
	}
	return nil
}

func (r repository) FindMany(query interface{}, conditions ...interface{}) ([]*Activity, restErrors.IRestErr) {
	var records []*Activity
	result := r.db.Where(query, conditions).Find(&records)
	if result.Error != nil {
		go logger.Error(repository.FindMany, result.Error)
		return nil, restErrors.NewInternalServerError("something went wrong")
	}
	return records, nil
}
