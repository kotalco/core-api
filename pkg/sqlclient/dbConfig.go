package sqlclient

import (
	"fmt"
	"github.com/kotalco/cloud-api/pkg/config"
	"github.com/kotalco/community-api/pkg/logger"
	"gorm.io/gorm"
	"strconv"
	"time"
)

func dbConfig(gormDbConnection *gorm.DB) {
	sqlDB, err := gormDbConnection.DB()
	if err != nil {
		go logger.Panic("DATABASE_CONNECTION_ERROR", err)
		panic(err)
	}

	databaseMaxConnections, err := strconv.Atoi(config.Environment.DatabaseMaxConnections)
	if err != nil {
		go logger.Panic("DATABASE_CONNECTION_ERROR", err)
		panic(err)
	}
	// Max Open Connection
	sqlDB.SetMaxOpenConns(databaseMaxConnections)

	databaseMaxIdleConnections, err := strconv.Atoi(config.Environment.DatabaseMaxIdleConnections)
	if err != nil {
		go logger.Panic("DATABASE_CONNECTION_ERROR", err)
		panic(err)
	}
	// Max Ideal Connection
	fmt.Println(databaseMaxIdleConnections)
	sqlDB.SetMaxIdleConns(databaseMaxIdleConnections)

	databaseMaxLifetimeConnections, err := strconv.Atoi(config.Environment.DatabaseMaxLifetimeConnections)
	if err != nil {
		go logger.Panic("DATABASE_CONNECTION_ERROR", err)
		panic(err)
	}
	// Connection Lifetime
	sqlDB.SetConnMaxLifetime(time.Duration(databaseMaxLifetimeConnections) * time.Minute)

	// Idle Connection Timeout
	databaseMaxIdleLifetimeConnections, err := strconv.Atoi(config.Environment.DatabaseMaxIdleLifetimeConnections)
	if err != nil {
		go logger.Panic("DATABASE_CONNECTION_ERROR", err)
		panic(err)
	}
	sqlDB.SetConnMaxIdleTime(time.Duration(databaseMaxIdleLifetimeConnections) * time.Minute)

}
