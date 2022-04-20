package sqlclient

import (
	"sync"

	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/cloud-api/pkg/config"
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
		dbConnection, err = gorm.Open(postgres.Open(config.EnvironmentConf["DB_SERVER_URL"]), &gorm.Config{})
		if err != nil {
			go logger.Panic("DATABASE_CONNECTION_ERROR", err)
			panic(err)
		}
		DbClient = dbConnection
	})
	DbClient = dbConnection

	return DbClient
}

func BeginTx() *gorm.DB {
	OpenDBConnection()
	return DbClient.Begin()
}

func RollbackTx(txHandle *gorm.DB) {
	txHandle.Rollback()
	OpenDBConnection()
}

func CommitTx(txHandle *gorm.DB) {
	txHandle.Commit()
	OpenDBConnection()
}
