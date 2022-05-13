package workspace

import (
	"github.com/google/uuid"
	"github.com/kotalco/cloud-api/internal/workspaceuser"
	"github.com/kotalco/cloud-api/pkg/security"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

var (
	repo = NewRepository()
)

func init() {
	err := sqlclient.OpenDBConnection().AutoMigrate(new(Workspace))
	if err != nil {
		panic(err.Error())
	}
	err = sqlclient.OpenDBConnection().AutoMigrate(new(workspaceuser.WorkspaceUser))
	if err != nil {
		panic(err.Error())
	}

}

func cleanUp(t *testing.T) {
	sqlclient.DbClient.Exec("TRUNCATE TABLE workspaces;")
	sqlclient.DbClient.Exec("TRUNCATE TABLE workspace_users;")
}

func TestRepository_Create(t *testing.T) {
	t.Run("Create_Should_Pass", func(t *testing.T) {
		createWorkspace(t)
		cleanUp(t)
	})

	t.Run("Create_Should_Throw_If_something_went_wrong_like_primary_id_duplication", func(t *testing.T) {
		workspace := createWorkspace(t)
		restErr := repo.Create(workspace)
		assert.EqualValues(t, "can't create workspace", restErr.Message)
		cleanUp(t)
	})
}

func TestRepository_GetByNameAndUserId(t *testing.T) {
	t.Run("Get_Workspace_By_Name_Should_Return_Workspace", func(t *testing.T) {
		workspace := createWorkspace(t)
		resp, err := repo.GetByNameAndUserId(workspace.Name, workspace.UserId)
		assert.Nil(t, err)
		assert.NotNil(t, resp)
		cleanUp(t)
	})

	t.Run("Get_Workspace_By_Name_Should_Throw_if_Record_Not_Found", func(t *testing.T) {
		resp, err := repo.GetByNameAndUserId("invalidName", "id")
		assert.Nil(t, resp)
		assert.EqualValues(t, http.StatusNotFound, err.Status)
		cleanUp(t)
	})
}

func createWorkspace(t *testing.T) *Workspace {
	workspace := new(Workspace)
	workspace.ID = uuid.New().String()
	workspace.UserId = uuid.New().String()
	workspace.Name = security.GenerateRandomString(10)
	workspace.K8sNamespace = workspace.ID
	restErr := repo.Create(workspace)
	assert.Nil(t, restErr)
	return workspace
}
