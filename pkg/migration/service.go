package migration

import (
	"gorm.io/gorm"
)

type service struct {
}

type IService interface {
	Migrate() []Definition
	Drop() []Definition
}

var migrations IMigration

func NewService(dbClient *gorm.DB) IService {
	migrations = NewMigration(dbClient)

	newService := new(service)
	return newService

}

func (service) Migrate() []Definition {
	return []Definition{
		Definition{
			Name: "CreateUserTable",
			Run: func() error {
				return migrations.CreateUserTable()
			},
		},
		Definition{
			Name: "CreateVerificationTable",
			Run: func() error {
				return migrations.CreateVerificationTable()
			},
		},
	}
}

func (service) Drop() []Definition {
	return []Definition{
		Definition{
			Name: "DropUserTable",
			Run: func() error {
				return migrations.DropUserTable()
			},
		},
		Definition{
			Name: "DropVerificationTable",
			Run: func() error {
				return migrations.DropVerificationTable()
			},
		},
	}
}
