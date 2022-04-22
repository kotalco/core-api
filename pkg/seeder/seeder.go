package seeder

import (
	"errors"
	"github.com/jackc/pgconn"
	"github.com/kotalco/cloud-api/internal/user"
	"github.com/kotalco/cloud-api/internal/verification"
	"gorm.io/gorm"
)

type Definition struct {
	Name string
	Run  func() error
}

type seeder struct {
	dbClient *gorm.DB
}

type ISeeder interface {
	SeedUserTable(users *user.User) error
	TruncateUserTable() error
	SeedVerificationTable(verifications *verification.Verification) error
	TruncateVerificationTable() error
}

func NewSeeder(dbClient *gorm.DB) ISeeder {
	newSeeder := new(seeder)
	newSeeder.dbClient = dbClient
	return newSeeder
}

func (s seeder) SeedUserTable(user *user.User) error {
	res := s.dbClient.FirstOrCreate(user)
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

func (s seeder) TruncateUserTable() error {
	res := s.dbClient.Exec("TRUNCATE TABLE users;")
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (s seeder) SeedVerificationTable(verification *verification.Verification) error {
	res := s.dbClient.FirstOrCreate(verification)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (s seeder) TruncateVerificationTable() error {
	res := s.dbClient.Exec("TRUNCATE TABLE verifications;")
	if res.Error != nil {
		return res.Error
	}
	return nil
}
