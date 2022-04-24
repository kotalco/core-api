package migration

import (
	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/cloud-api/internal/user"
	"github.com/kotalco/cloud-api/internal/verification"
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
}

func NewMigration(dbClient *gorm.DB) IMigration {
	newMigration := new(migration)
	newMigration.dbClient = dbClient
	return newMigration
}

func (m migration) CreateUserTable() error {
	exits := m.dbClient.Migrator().HasTable(user.User{})
	if !exits {
		go logger.Info("CreateUserTable")
		return m.dbClient.AutoMigrate(user.User{})
	}
	return nil
}

func (m migration) CreateVerificationTable() error {
	exits := m.dbClient.Migrator().HasTable(verification.Verification{})
	if !exits {
		go logger.Info("CreateVerificationTable")
		return m.dbClient.AutoMigrate(verification.Verification{})
	}
	return nil
}
