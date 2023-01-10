package psql

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/subscriptions-api/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

type Config struct {
	DBServerURL string
}

func New(config Config) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) (checkErr error) {

		dbConnection, err := gorm.Open(postgres.Open(config.DBServerURL), &gorm.Config{
			Logger: glogger.Default.LogMode(glogger.Error),
		})
		if err != nil {
			checkErr = fmt.Errorf("PostgreSQL health check failed on connect: %w", err)
			go logger.Error("PSQL_HEALTH_CHECK", err)
			return
		}

		dbInstance, err := dbConnection.DB()
		if err != nil {
			checkErr = fmt.Errorf("PostgreSQL health check failed on gettting db instance")
			go logger.Error("PSQL_HEALTH_CHECK", err)
			return
		}

		defer func() {
			if err = dbInstance.Close(); err != nil && checkErr == nil {
				checkErr = fmt.Errorf("PostgreSQL health check failed on connection closing: %w", err)
				go logger.Error("DATABASE_CONNECTION_ERROR", err)
			}
		}()

		err = dbInstance.PingContext(ctx.Context())
		if err != nil {
			go logger.Error("DATABASE_CONNECTION_ERROR", err)
			checkErr = fmt.Errorf("PostgreSQL health check failed on ping: %w", err)
			return
		}

		rows, err := dbInstance.QueryContext(ctx.Context(), `SELECT VERSION()`)
		if err != nil {
			go logger.Error("DATABASE_CONNECTION_ERROR", err)
			checkErr = fmt.Errorf("PostgreSQL health check failed on select: %w", err)
			return
		}
		defer func() {
			if err = rows.Close(); err != nil && checkErr == nil {
				checkErr = fmt.Errorf("PostgreSQL health check failed on rows closing: %w", err)
			}
		}()

		return
	}
}
