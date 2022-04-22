package seeder

import (
	"github.com/google/uuid"
	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/cloud-api/internal/user"
	"github.com/kotalco/cloud-api/pkg/security"
	"gorm.io/gorm"
)

type service struct{}

type IService interface {
	Seed() []Definition
	Truncate() []Definition
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

func (s service) Seed() []Definition {
	return []Definition{
		Definition{
			Name: "SeedUsersTable",
			Run: func() error {
				hashedPassword, err := hashing.Hash("develop", 13)
				if err != nil {
					go logger.Error("SEEDING_HASHING_ERR", err)
				}
				developUser := new(user.User)
				developUser.ID = uuid.New().String()
				developUser.Email = "develop@kotal.co"
				developUser.IsEmailVerified = true
				developUser.Password = string(hashedPassword)
				return seeders.SeedUserTable(developUser)
			},
		},
	}
}

func (s service) Truncate() []Definition {
	return []Definition{
		Definition{
			Name: "TruncateUserTable",
			Run: func() error {
				return seeders.TruncateUserTable()
			},
		},
		Definition{
			Name: "TruncateVerificationTable",
			Run: func() error {
				return seeders.TruncateVerificationTable()
			},
		},
	}
}
