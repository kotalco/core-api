package sqlclient

import (
	"sync"

	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/cloud-api/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DbClient   *gorm.DB
	clientOnce sync.Once
)

func OpenDBConnection() *gorm.DB {
	clientOnce.Do(func() {
		db, err := gorm.Open(postgres.Open(config.EnvironmentConf["DB_SERVER_URL"]), &gorm.Config{})
		if err != nil {
			go logger.Panic("DATABASE_CONNECTION_ERROR", err)
			panic(err)
		}
		DbClient = db
	})

	return DbClient
}
