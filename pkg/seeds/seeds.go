package seeds

import (
	"fmt"
	"github.com/google/uuid"
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
			Name: "CreateDefaultUser",
			Run: func(db *gorm.DB) error {
				hashedPassword, err := hashing.Hash("123456", 13)
				if err != nil {

				}
				user1 := new(user.User)
				user1.ID = uuid.New().String()
				user1.Email = security.GenerateRandomString(5) + "@gmail.com"
				user1.IsEmailVerified = true
				user1.Password = string(hashedPassword)
				fmt.Println("Default user", "Email:", user1.Email, ",  Password: ", "123456")

				return CreateUser(db, user1)
			},
		},
	}
}

func ClearDB(dbClient *gorm.DB) {
	dbClient.Exec("TRUNCATE TABLE users;")
	dbClient.Exec("TRUNCATE TABLE verifications;")
}
