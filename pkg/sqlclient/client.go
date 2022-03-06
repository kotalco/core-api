package sqlclient

import (
	"fmt"
	"github.com/kotalco/cloud-api/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"sync"
)

var DbClient *gorm.DB

var lock = &sync.Mutex{}

func OpenDBConnection() (*gorm.DB, error) {
	lock.Lock()
	defer lock.Unlock()
	if DbClient == nil {
		db, err := gorm.Open(postgres.Open(config.EnvironmentConf["DB_SERVER_URL"]), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("error, not connected to database, %w", err)
		}
		DbClient = db
	}
	return DbClient, nil
}
