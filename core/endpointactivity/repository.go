package endpointactivity

import (
	"github.com/kotalco/cloud-api/config"
	restErrors "github.com/kotalco/cloud-api/pkg/errors"
	"github.com/kotalco/cloud-api/pkg/logger"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	"gorm.io/gorm"
	"strconv"
)

var (
	rawDailyStatsQuery  = "SELECT EXTRACT(DAY FROM timestamp) AS day, COUNT(*) AS count FROM activities WHERE endpoint_id = $1 AND timestamp >= $2 AND timestamp <= $3 GROUP BY day ORDER BY day"
	rawWeeklyStatsQuery = "SELECT EXTRACT(WEEK FROM timestamp) AS week, COUNT(*) AS count FROM activities WHERE endpoint_id = $1 AND timestamp >= $2 AND timestamp <= $3 GROUP BY week ORDER BY week;"
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
