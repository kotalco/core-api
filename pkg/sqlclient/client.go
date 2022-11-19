package sqlclient

import (
	"sync"

	glogger "gorm.io/gorm/logger"

	"github.com/kotalco/cloud-api/pkg/config"
	"github.com/kotalco/community-api/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DbClient     *gorm.DB
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
			go logger.Panic("DATABASE_CONNECTION_ERROR", err)
			panic(err)
		}
		dbConfig(dbConnection)
		DbClient = dbConnection
	})
	DbClient = dbConnection
	return dbConnection
}

func Begin() *gorm.DB {
	DbClient = dbConnection
	return DbClient.Begin()
}

func Rollback(txHandle *gorm.DB) {
	txHandle.Rollback()
	DbClient = dbConnection
}

func Commit(txHandle *gorm.DB) {
	txHandle.Commit()
	DbClient = dbConnection
}
