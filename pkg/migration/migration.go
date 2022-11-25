package migration

import (
	"github.com/kotalco/cloud-api/internal/user"
	"github.com/kotalco/cloud-api/internal/verification"
	"github.com/kotalco/cloud-api/internal/workspace"
	"github.com/kotalco/cloud-api/internal/workspaceuser"
	"github.com/kotalco/community-api/pkg/logger"
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
	go logger.Info("CreateUserTable")
	return nil
}

func (m migration) CreateVerificationTable() error {
	err := m.dbClient.Migrator().AutoMigrate(verification.Verification{})
	if err != nil {
		go logger.Error(m.CreateVerificationTable, err)
		return err
	}
	go logger.Info("CreateVerificationTable")
	return nil
}

func (m migration) CreateWorkspaceTable() error {
	err := m.dbClient.Migrator().AutoMigrate(workspace.Workspace{})
	if err != nil {
		go logger.Error(m.CreateWorkspaceUserTable, err)
		return err
	}
	go logger.Info("CreateWorkspaceTable")
	return nil
}

func (m migration) CreateWorkspaceUserTable() error {
	err := m.dbClient.Migrator().AutoMigrate(workspaceuser.WorkspaceUser{})
	if err != nil {
		go logger.Error(m.CreateWorkspaceUserTable, err)
		return err
	}
	go logger.Info("CreateWorkspaceUserTable")
	return nil
}
