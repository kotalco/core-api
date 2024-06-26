package migration

import (
	"github.com/kotalco/core-api/core/endpointactivity"
	"github.com/kotalco/core-api/core/setting"
	"github.com/kotalco/core-api/core/user"
	"github.com/kotalco/core-api/core/verification"
	"github.com/kotalco/core-api/core/workspace"
	"github.com/kotalco/core-api/core/workspaceuser"
	"github.com/kotalco/core-api/pkg/logger"
	"gorm.io/gorm"
)

type Definition struct {
	Name string
	Run  func() error
}

type migration struct {
	dbClient *gorm.DB
}

type IMigration interface {
	CreateUserTable() error
	CreateVerificationTable() error
	CreateWorkspaceTable() error
	CreateWorkspaceUserTable() error
	CreateSettingTable() error
	CreateEndpointActivityTable() error
}

func NewMigration(dbClient *gorm.DB) IMigration {
	newMigration := new(migration)
	newMigration.dbClient = dbClient
	return newMigration
}

func (m migration) CreateUserTable() error {
	err := m.dbClient.Migrator().AutoMigrate(user.User{})
	if err != nil {
		go logger.Error(m.CreateWorkspaceTable, err)
		return err
	}
	go logger.Info(m.CreateUserTable, "CreateUserTable")
	return nil
}

func (m migration) CreateVerificationTable() error {
	err := m.dbClient.Migrator().AutoMigrate(verification.Verification{})
	if err != nil {
		go logger.Error(m.CreateVerificationTable, err)
		return err
	}
	go logger.Info(m.CreateVerificationTable, "CreateVerificationTable")
	return nil
}

func (m migration) CreateWorkspaceTable() error {
	err := m.dbClient.Migrator().AutoMigrate(workspace.Workspace{})
	if err != nil {
		go logger.Error(m.CreateWorkspaceUserTable, err)
		return err
	}
	go logger.Info(m.CreateWorkspaceTable, "CreateWorkspaceTable")
	return nil
}

func (m migration) CreateWorkspaceUserTable() error {
	err := m.dbClient.Migrator().AutoMigrate(workspaceuser.WorkspaceUser{})
	if err != nil {
		go logger.Error(m.CreateWorkspaceUserTable, err)
		return err
	}
	go logger.Info(m.CreateWorkspaceUserTable, "CreateWorkspaceUserTable")
	return nil
}

func (m migration) CreateSettingTable() error {
	exits := m.dbClient.Migrator().HasTable(setting.Setting{})
	if !exits {
		go logger.Info(m.CreateSettingTable, "CreateSettingTable")
		return m.dbClient.AutoMigrate(setting.Setting{})
	}
	return nil
}

func (m migration) CreateEndpointActivityTable() error {
	exits := m.dbClient.Migrator().HasTable(endpointactivity.Activity{})
	if !exits {
		go logger.Info(m.CreateEndpointActivityTable, "CreateEndpointActivityTable")
		return m.dbClient.AutoMigrate(endpointactivity.Activity{})
	}
	return nil
}
