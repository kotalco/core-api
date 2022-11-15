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

func OpenDBConnection() *gorm.DB {
	clientOnce.Do(func() {
		dbConnection, err = gorm.Open(postgres.Open(config.Environment.DatabaseServerURL), &gorm.Config{
			Logger: glogger.Default.LogMode(glogger.Error),
		})
		if err != nil {
			go logger.Panic("DATABASE_CONNECTION_ERROR", err)
			panic(err)
		}
		DbClient = dbConnection
	})
	DbClient = dbConnection

	return DbClient
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
