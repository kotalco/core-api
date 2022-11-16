package migration

import (
	"gorm.io/gorm"
)

type service struct {
}

type IService interface {
	Migrations() []Definition
}

var migrator IMigration

func NewService(dbClient *gorm.DB) IService {
	migrator = NewMigration(dbClient)

	newService := new(service)
	return newService

}

func (service) Migrations() []Definition {
	return []Definition{
		Definition{
			Name: "CreateUserTable",
			Run: func() error {
				return migrator.CreateUserTable()
			},
		},
		Definition{
			Name: "CreateVerificationTable",
			Run: func() error {
				return migrator.CreateVerificationTable()
			},
		},
		Definition{
			Name: "CreateWorkspaceTable",
			Run: func() error {
				return migrator.CreateWorkspaceTable()
			},
		},
		Definition{Name: "CreateWorkspaceUserTable",
			Run: func() error {
				return migrator.CreateWorkspaceUserTable()
			},
		},
		Definition{Name: "CreateDbConfigTable",
			Run: func() error {
				return migrator.CreateDbConfigTable()
			},
		},
	}
}
