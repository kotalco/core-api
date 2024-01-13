package migration

import (
	"github.com/kotalco/cloud-api/pkg/logger"
	"gorm.io/gorm"
)

const (
	MigrateUserTable             = "MigrateUserTable"
	MigrateVerificationTable     = "MigrateVerificationTable"
	MigrateWorkspaceTable        = "MigrateWorkspaceTable"
	MigrateWorkspaceUserTable    = "MigrateWorkspaceUserTable"
	MigrateSettingTable          = "MigrateSettingTable"
	MigrateEndpointActivityTable = "MigrateEndpointActivityTable"
)

type service struct {
}

type IService interface {
	Migrations() map[string]Definition
	Run()
}

var migrator IMigration

func NewService(dbClient *gorm.DB) IService {
	migrator = NewMigration(dbClient)

	newService := new(service)
	return newService

}

func (service) Migrations() map[string]Definition {
	return map[string]Definition{
		MigrateUserTable: {
			Name: MigrateUserTable,
			Run: func() error {
				return migrator.CreateUserTable()
			},
		},
		MigrateVerificationTable: {
			Name: MigrateVerificationTable,
			Run: func() error {
				return migrator.CreateVerificationTable()
			},
		},
		MigrateWorkspaceTable: {
			Name: MigrateWorkspaceTable,
			Run: func() error {
				return migrator.CreateWorkspaceTable()
			},
		},
		MigrateWorkspaceUserTable: {
			Name: MigrateWorkspaceUserTable,
			Run: func() error {
				return migrator.CreateWorkspaceUserTable()
			},
		},
		MigrateSettingTable: {
			Name: MigrateSettingTable,
			Run: func() error {
				return migrator.CreateSettingTable()
			},
		},
		MigrateEndpointActivityTable: {
			Name: MigrateEndpointActivityTable,
			Run: func() error {
				return migrator.CreateEndpointActivityTable()
			},
		},
	}
}

func (s service) Run() {
	for _, step := range s.Migrations() {
		if err := step.Run(); err != nil {
			go logger.Error(step.Name, err)
		}
	}
}
