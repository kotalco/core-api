package workspaceuser

import (
	"github.com/google/uuid"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	repo = NewRepository()
)

func init() {
	err := sqlclient.OpenDBConnection().AutoMigrate(new(WorkspaceUser))
	if err != nil {
		panic(err.Error())
	}

}

func cleanUp(workspaceUser WorkspaceUser) {
	sqlclient.DbClient.Delete(workspaceUser)
}

func TestRepository_Create(t *testing.T) {
	t.Run("Create_Should_Pass", func(t *testing.T) {
		model := createWorkspaceUser(t)
		cleanUp(model)
	})
}

func createWorkspaceUser(t *testing.T) WorkspaceUser {
	model := new(WorkspaceUser)
	model.ID = uuid.New().String()
	model.UserId = uuid.New().String()
	model.WorkspaceID = uuid.New().String()
	restErr := repo.Create(model)
	assert.Nil(t, restErr)
	return *model
}
