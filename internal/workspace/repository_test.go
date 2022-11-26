package workspace

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/kotalco/cloud-api/internal/workspaceuser"
	"github.com/kotalco/cloud-api/pkg/roles"
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
	err := sqlclient.OpenDBConnection().AutoMigrate(new(Workspace), new(workspaceuser.WorkspaceUser))
	if err != nil {
		panic(err.Error())
	}
}

func cleanUp(workspace Workspace) {
	sqlclient.OpenDBConnection().Delete(workspace)
}

func TestRepository_Create(t *testing.T) {
	t.Run("Create_Should_Pass", func(t *testing.T) {
		workspace := createWorkspace(t)
		cleanUp(workspace)
	})

	t.Run("Create_Should_Throw_If_something_went_wrong_like_primary_id_duplication", func(t *testing.T) {
		workspace := createWorkspace(t)
		restErr := repo.Create(&workspace)
		assert.EqualValues(t, "can't create workspace", restErr.Message)
		cleanUp(workspace)
	})
}

func TestWorkspaceRepository_Update(t *testing.T) {
	t.Run("update_should_pass", func(t *testing.T) {
		workspace := createWorkspace(t)
		workspace.Name = "newname"
		err := repo.Update(&workspace)
		assert.Nil(t, err)
		cleanUp(workspace)
	})
}

func TestRepository_GetByNameAndUserId(t *testing.T) {
	t.Run("Get_Workspace_By_Name_Should_Return_Workspace", func(t *testing.T) {
		workspace := createWorkspace(t)
		resp, err := repo.GetByNameAndUserId(workspace.Name, workspace.UserId)
		fmt.Println(resp, err)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, len(resp))
		cleanUp(workspace)
	})

	t.Run("Get_Workspace_By_Name_Should_return_empty_slice", func(t *testing.T) {
		resp, err := repo.GetByNameAndUserId("invalidName", "id")
		assert.Nil(t, err)
		assert.EqualValues(t, 0, len(resp))
	})
}

func TestRepository_GetById(t *testing.T) {
	t.Run("Get_Workspace_By_Id_Should_Return_Workspace", func(t *testing.T) {
		workspace := createWorkspace(t)
		resp, err := repo.GetById(workspace.ID)
		fmt.Println(workspace.WorkspaceUsers)
		assert.Nil(t, err)
		assert.NotNil(t, resp)
		cleanUp(workspace)
	})

	t.Run("Get_Workspace_By_Name_Should_Throw_if_Record_Not_Found", func(t *testing.T) {
		resp, err := repo.GetById("invalidName")
		assert.Nil(t, resp)
		assert.EqualValues(t, http.StatusNotFound, err.Status)
	})
}

func TestRepository_Delete(t *testing.T) {
	t.Run("Delete_work_space_should_pass", func(t *testing.T) {
		workspace := createWorkspace(t)
		err := repo.Delete(&workspace)
		assert.Nil(t, err)
		model, err := repo.GetById(workspace.ID)
		assert.Nil(t, model)
		assert.Error(t, err)
	})

	t.Run("Delete_Workspace_should_throw_not_found_if_record_already_deleted", func(t *testing.T) {
		workspace := new(Workspace)
		workspace.ID = "invalid"
		err := repo.Delete(workspace)
		assert.EqualValues(t, http.StatusNotFound, err.Status)
	})
}

func TestRepository_GetByUserId(t *testing.T) {
	t.Run("Get_Workspaces_By_UserId_Should_Return_Workspace", func(t *testing.T) {
		var list []*Workspace
		workspace := createWorkspace(t)
		list = append(list, &workspace)
		resp, err := repo.GetByUserId(list[0].UserId)
		assert.Nil(t, err)
		assert.NotNil(t, resp)
		cleanUp(*list[0])
	})

	t.Run("Get_Workspace_By_Name_Should_return_empty_list_when_is_invalid", func(t *testing.T) {
		resp, err := repo.GetByUserId("invalidName")
		assert.Nil(t, err)
		var list = make([]*Workspace, 0)
		assert.EqualValues(t, list, resp)
	})
}

func TestRepository_AddWorkspaceMember(t *testing.T) {
	t.Run("Add_workspace_member_should_pass", func(t *testing.T) {
		workspace := createWorkspace(t)
		newRecord := new(workspaceuser.WorkspaceUser)
		newRecord.ID = uuid.NewString()
		newRecord.WorkspaceID = workspace.ID
		newRecord.UserId = workspace.UserId
		err := repo.AddWorkspaceMember(&workspace, newRecord)
		assert.Nil(t, err)

		model, err := repo.GetById(workspace.ID)
		assert.Nil(t, err)
		assert.EqualValues(t, 2, len(model.WorkspaceUsers))

		cleanUp(workspace)
	})
}

func TestRepository_DeleteWorkspaceMember(t *testing.T) {
	t.Run("delete_workspace_member_should_pass", func(t *testing.T) {
		workspace := createWorkspace(t)
		err := repo.DeleteWorkspaceMember(&workspace, &workspace.WorkspaceUsers[0])
		assert.Nil(t, err)

		model, err := repo.GetById(workspace.ID)
		assert.Nil(t, err)
		assert.EqualValues(t, 0, len(model.WorkspaceUsers))

		cleanUp(workspace)
	})
}

func TestRepository_GetWorkspaceMemberByWorkspaceIdAndUserId(t *testing.T) {
	t.Run("get_workspace_member_by_workspace_id_and_user_id_should_pass", func(t *testing.T) {
		workspace := createWorkspace(t)
		workspaceUser, err := repo.GetWorkspaceMemberByWorkspaceIdAndUserId(workspace.ID, workspace.WorkspaceUsers[0].UserId)

		assert.Nil(t, err)
		assert.EqualValues(t, workspace.WorkspaceUsers[0], *workspaceUser)

		cleanUp(workspace)
	})
}

func TestRepository_CountByUserId(t *testing.T) {
	t.Run("count_workspace_by_user_id_should_pass", func(t *testing.T) {
		workspace := createWorkspace(t)
		count, err := repo.CountByUserId(workspace.UserId)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, count)
		cleanUp(workspace)

	})
}

func TestRepository_UpdateWorkspaceUser(t *testing.T) {
	t.Run("update_workspace_member_should_pass", func(t *testing.T) {
		workspace := createWorkspace(t)
		err := repo.UpdateWorkspaceUser(&workspace.WorkspaceUsers[0])
		assert.Nil(t, err)
		cleanUp(workspace)

	})
}

func TestRepository_GetByNamespace(t *testing.T) {
	t.Run("getByNamespace_should_pass", func(t *testing.T) {
		workspace := createWorkspace(t)
		model, err := repo.GetByNamespace(workspace.K8sNamespace)
		assert.Nil(t, err)
		assert.NotNil(t, model)
		cleanUp(workspace)
	})
}

func createWorkspace(t *testing.T) Workspace {
	workspace := new(Workspace)
	workspace.ID = uuid.New().String()
	workspace.UserId = uuid.New().String()
	workspace.Name = security.GenerateRandomString(10)
	workspace.K8sNamespace = workspace.ID

	newWorkspaceUser := new(workspaceuser.WorkspaceUser)
	newWorkspaceUser.ID = uuid.New().String()
	newWorkspaceUser.WorkspaceID = workspace.ID
	newWorkspaceUser.UserId = workspace.UserId
	newWorkspaceUser.Role = roles.Admin
	workspace.WorkspaceUsers = append(workspace.WorkspaceUsers, *newWorkspaceUser)
	restErr := repo.Create(workspace)
	assert.Nil(t, restErr)
	return *workspace
}
