package seeder

import (
	"errors"
	"github.com/jackc/pgconn"
	"github.com/kotalco/cloud-api/core/setting"
	"gorm.io/gorm"
)

type Definition struct {
	Run func() error
}

type seeder struct {
	dbClient *gorm.DB
}

type ISeeder interface {
	SeedSettingTable(setting *setting.Setting) error
}

func NewSeeder(dbClient *gorm.DB) ISeeder {
	newSeeder := new(seeder)
	newSeeder.dbClient = dbClient
	return newSeeder
}

func (s seeder) SeedSettingTable(setting *setting.Setting) error {
	res := s.dbClient.Create(setting)
	if res.Error != nil {
		var pgErr *pgconn.PgError
		if errors.As(res.Error, &pgErr) {
			if pgErr.Code != "23505" {
				return res.Error
			}
		}
	}
	return nil
}
