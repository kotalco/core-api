package endpointactivity

import (
	"github.com/kotalco/cloud-api/pkg/config"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	"gorm.io/gorm"
	"strconv"
)

var (
	rawStatsQuery = `SELECT
	  COUNT(id) FILTER (WHERE date_trunc('month', to_timestamp(timestamp)) = date_trunc('month', current_date)) AS monthly_hits,
	  COUNT(id) FILTER (WHERE date_trunc('week', to_timestamp(timestamp)) = date_trunc('week', current_date)) AS weekly_hits,
	  COUNT(id) FILTER (WHERE date_trunc('day', to_timestamp(timestamp)) = date_trunc('day', current_date)) AS daily_hits
	FROM
	  activities
	WHERE
	  endpoint_id = ?`
)

type repository struct {
	db *gorm.DB
}

type IRepository interface {
	WithTransaction(txHandle *gorm.DB) IRepository
	WithoutTransaction() IRepository
	CreateInBatches(activities []*Activity) restErrors.IRestErr
	RawQuery(query string, dest interface{}, conditions ...interface{}) restErrors.IRestErr
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
	batchSize, err := strconv.Atoi(config.Environment.DatabaseInsertBatchSize)
	if err != nil {
		logger.Warn("CreateInBatches", err)
		batchSize = len(activities)
	}

	res := r.db.CreateInBatches(activities, batchSize)
	if res.Error != nil {
		go logger.Error(r.CreateInBatches, res.Error)
		return restErrors.NewInternalServerError("something went wrong")
	}
	return nil
}

func (r repository) RawQuery(query string, dest interface{}, conditions ...interface{}) restErrors.IRestErr {
	result := r.db.Raw(query, conditions...).Scan(dest)
	if result.Error != nil {
		go logger.Error(repository.RawQuery, result.Error)
		return restErrors.NewInternalServerError("something went wrong")
	}
	return nil
}
