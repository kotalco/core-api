package seeder

import (
	"github.com/kotalco/cloud-api/internal/setting"
	"github.com/kotalco/cloud-api/pkg/security"
	"github.com/kotalco/community-api/pkg/logger"
	"gorm.io/gorm"
)

const (
	SeedUsersTable     = "SeedUsersTable"
	SeedWorkspaceTable = "SeedWorkspaceTable"
	SeedSettingTable   = "SeedSettingTable"
)

type service struct {
}

type IService interface {
	Seeds() map[string]Definition
	Run()
}

var (
	seeders ISeeder
	hashing = security.NewHashing()
)

func NewService(dbClient *gorm.DB) IService {
	seeders = NewSeeder(dbClient)
	newService := new(service)
	return newService
}

func (s service) Seeds() map[string]Definition {
	return map[string]Definition{
		SeedSettingTable: {
			Run: func() error {
				record := new(setting.Setting)
				record.Key = setting.RegistrationKey
				record.Value = "true"
				return seeders.SeedSettingTable(record)
			},
		},
	}
}

func (s service) Run() {
	//seed setting table
	err := s.Seeds()[SeedSettingTable].Run()
	if err != nil {
		go logger.Error(seeder.SeedSettingTable, err)
	}
}
