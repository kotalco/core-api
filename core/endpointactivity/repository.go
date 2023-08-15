package endpointactivity

import (
	"fmt"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

type IRepository interface {
	WithTransaction(txHandle *gorm.DB) IRepository
	WithoutTransaction() IRepository
	GetByEndpointId(endpointId string) (*Activity, restErrors.IRestErr)
	increment(activity *Activity) restErrors.IRestErr
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

func (r repository) GetByEndpointId(endpointId string) (*Activity, restErrors.IRestErr) {
	var activity = new(Activity)

	result := r.db.Where("endpoint_id = ?", endpointId).First(activity)
	if result.Error != nil {
		return nil, restErrors.NewNotFoundError(fmt.Sprintf("can't find activity with endpointId  %s", endpointId))
	}

	return activity, nil
}

func (r repository) increment(activity *Activity) restErrors.IRestErr {
	res := r.db.Save(activity)
	if res.Error != nil {
		go logger.Error(r.increment, res.Error)
		return restErrors.NewInternalServerError("something went wrong")
	}
	return nil
}
