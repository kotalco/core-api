package seeder

import (
	"github.com/google/uuid"
	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/cloud-api/internal/user"
	"github.com/kotalco/cloud-api/internal/workspace"
	"github.com/kotalco/cloud-api/internal/workspaceuser"
	"github.com/kotalco/cloud-api/pkg/roles"
	"github.com/kotalco/cloud-api/pkg/security"
	"gorm.io/gorm"
)

type service struct{}

type IService interface {
	Seeds() []Definition
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

func (s service) Seeds() []Definition {
	defaultUserId := uuid.NewString()
	return []Definition{
		Definition{
			Name: "SeedUsersTable",
			Run: func() error {
				hashedPassword, err := hashing.Hash("develop", 13)
				if err != nil {
					go logger.Error("SEEDING_HASHING_ERR", err)
				}
				developUser := new(user.User)
				developUser.ID = defaultUserId
				developUser.Email = "develop@kotal.co"
				developUser.IsEmailVerified = true
				developUser.Password = string(hashedPassword)
				return seeders.SeedUserTable(developUser)
			},
		},
		Definition{
			Name: "SeedWorkspaceTable",
			Run: func() error {
				newWorkspace := new(workspace.Workspace)
				newWorkspace.ID = uuid.New().String()
				newWorkspace.UserId = defaultUserId
				newWorkspace.Name = "default"
				newWorkspace.K8sNamespace = "default"
				//create workspace-user record
				newWorkspaceuser := new(workspaceuser.WorkspaceUser)
				newWorkspaceuser.ID = uuid.New().String()
				newWorkspaceuser.WorkspaceID = newWorkspace.ID
				newWorkspaceuser.UserId = newWorkspace.UserId
				newWorkspaceuser.Role = roles.Admin

				newWorkspace.WorkspaceUsers = append(newWorkspace.WorkspaceUsers, *newWorkspaceuser)

				return seeders.SeedWorkspaceTable(newWorkspace)
			},
		},
	}
}
