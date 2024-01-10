package sqlclient

import (
	"database/sql"
	"github.com/kotalco/cloud-api/config"
	"sync"

	glogger "gorm.io/gorm/logger"

	"github.com/kotalco/cloud-api/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	dbConnection *gorm.DB
	clientOnce   sync.Once
	err          error
)

// OpenDBConnection when initiating new repository we should use this function
// we use OpenDBConnection which returns the original dbConnection variable from sqlclient pkg
// don't use sqlclinet.DbClient might have been polluted with another transaction
func OpenDBConnection() *gorm.DB {
	clientOnce.Do(func() {
		dbConnection, err = gorm.Open(postgres.Open(config.Environment.DatabaseServerURL), &gorm.Config{
			Logger: glogger.Default.LogMode(glogger.Error),
		})
		if err != nil {
			go logger.Warn("DATABASE_CONNECTION_ERROR", err)
			panic(err)
		}
		dbConfig(dbConnection)
	})
	return dbConnection
}

func Begin(opts ...*sql.TxOptions) *gorm.DB {
	return dbConnection.Begin(opts...)

}

func Rollback(txHandle *gorm.DB) {
	txHandle.Rollback()
}

func Commit(txHandle *gorm.DB) {
	txHandle.Commit()
}
