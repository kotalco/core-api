package seeder

import (
	"github.com/google/uuid"
	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/cloud-api/internal/user"
	"github.com/kotalco/cloud-api/pkg/security"
	"gorm.io/gorm"
)

type Seed struct {
	Name string
	Run  func(db *gorm.DB) error
}

var hashing = security.NewHashing()

func All() []Seed {
	return []Seed{
		Seed{
			Name: "CreateDevelopUser",
			Run: func(db *gorm.DB) error {
				db.AutoMigrate(user.User{})
				hashedPassword, err := hashing.Hash("develop", 13)
				if err != nil {
					go logger.Error("SEEDING_ERR", err)
				}
				developUser := new(user.User)
				developUser.ID = uuid.New().String()
				developUser.Email = "develop@kotal.co"
				developUser.IsEmailVerified = true
				developUser.Password = string(hashedPassword)
				go logger.Info("develop user has been seeded")
				return CreateUser(db, developUser)
			},
		},
	}
}

func ClearDB(dbClient *gorm.DB) {
	dbClient.Exec("TRUNCATE TABLE users;")
	dbClient.Exec("TRUNCATE TABLE verifications;")
}
