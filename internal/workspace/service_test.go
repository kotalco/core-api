package workspace

import (
	"github.com/google/uuid"
	"github.com/kotalco/cloud-api/internal/workspaceuser"
	"github.com/kotalco/cloud-api/pkg/roles"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"net/http"
	"os"
	"testing"
)

var (
	WithTransactionFunc                          func(txHandle *gorm.DB) IRepository
	workspaceTestService                         IService
	CreateWorkspaceFunc                          func(workspace *Workspace) *restErrors.RestErr
	UpdateWorkspaceFunc                          func(workspace *Workspace) *restErrors.RestErr
	GetByNameAndUserIdFunc                       func(name string, userId string) ([]*Workspace, *restErrors.RestErr)
	GetByIdFunc                                  func(Id string) (*Workspace, *restErrors.RestErr)
	DeleteFunc                                   func(workspace *Workspace) *restErrors.RestErr
	GetByUserIdFunc                              func(userId string) ([]*Workspace, *restErrors.RestErr)
	addWorkspaceMemberFunc                       func(workspace *Workspace, workspaceUser *workspaceuser.WorkspaceUser) *restErrors.RestErr
	DeleteWorkspaceMemberFunc                    func(workspace *Workspace, workspaceUser *workspaceuser.WorkspaceUser) *restErrors.RestErr
	GetWorkspaceMemberByWorkspaceIdAndUserIdFunc func(workspaceId string, userId string) (*workspaceuser.WorkspaceUser, *restErrors.RestErr)
	CountByUserIdFunc                            func(userId string) (int64, *restErrors.RestErr)
	UpdateWorkspaceUserFunc                      func(workspaceUser *workspaceuser.WorkspaceUser) *restErrors.RestErr
	GetByNamespaceFunc                           func(namespace string) (*Workspace, *restErrors.RestErr)
)

type workspaceRepositoryMock struct{}

func (r workspaceRepositoryMock) UpdateWorkspaceUser(workspaceUser *workspaceuser.WorkspaceUser) *restErrors.RestErr {
	return UpdateWorkspaceUserFunc(workspaceUser)
}

func (r workspaceRepositoryMock) WithTransaction(txHandle *gorm.DB) IRepository {
	return r
}

func (workspaceRepositoryMock) Create(workspace *Workspace) *restErrors.RestErr {
	return CreateWorkspaceFunc(workspace)
}

func (workspaceRepositoryMock) Update(workspace *Workspace) *restErrors.RestErr {
	return UpdateWorkspaceFunc(workspace)
}

func (workspaceRepositoryMock) GetByNameAndUserId(name string, userId string) ([]*Workspace, *restErrors.RestErr) {
	return GetByNameAndUserIdFunc(name, userId)
}
func (workspaceRepositoryMock) GetById(Id string) (*Workspace, *restErrors.RestErr) {
	return GetByIdFunc(Id)
}

func (workspaceRepositoryMock) Delete(workspace *Workspace) *restErrors.RestErr {
	return DeleteFunc(workspace)
}

func (workspaceRepositoryMock) GetByUserId(userId string) ([]*Workspace, *restErrors.RestErr) {
	return GetByUserIdFunc(userId)
}

func (workspaceRepositoryMock) AddWorkspaceMember(workspace *Workspace, workspaceUser *workspaceuser.WorkspaceUser) *restErrors.RestErr {
	return addWorkspaceMemberFunc(workspace, workspaceUser)
}

func (workspaceRepositoryMock) DeleteWorkspaceMember(workspace *Workspace, workspaceUser *workspaceuser.WorkspaceUser) *restErrors.RestErr {
	return DeleteWorkspaceMemberFunc(workspace, workspaceUser)
}
func (workspaceRepositoryMock) GetWorkspaceMemberByWorkspaceIdAndUserId(workspaceId string, userId string) (*workspaceuser.WorkspaceUser, *restErrors.RestErr) {
	return GetWorkspaceMemberByWorkspaceIdAndUserIdFunc(workspaceId, userId)
}
func (workspaceRepositoryMock) CountByUserId(userId string) (int64, *restErrors.RestErr) {
	return CountByUserIdFunc(userId)
}

func (workspaceRepositoryMock) GetByNamespace(namespace string) (*Workspace, *restErrors.RestErr) {
	return GetByNamespaceFunc(namespace)
}

func TestMain(m *testing.M) {
	workspaceRepo = &workspaceRepositoryMock{}
	workspaceTestService = NewService()
	code := m.Run()
	os.Exit(code)
}

func TestService_Create(t *testing.T) {
	dto := new(CreateWorkspaceRequestDto)
	dto.Name = "testName"
	t.Run("Create_Workspace_Should_Pass", func(t *testing.T) {
		GetByNameAndUserIdFunc = func(name string, userId string) ([]*Workspace, *restErrors.RestErr) {
			return []*Workspace{}, nil
		}
		CreateWorkspaceFunc = func(workspace *Workspace) *restErrors.RestErr {
			return nil
		}

		model, err := workspaceTestService.Create(dto, "1")
		assert.Nil(t, err)
		assert.NotNil(t, model)
		assert.EqualValues(t, roles.Admin, model.WorkspaceUsers[0].Role)
	})

	t.Run("WorkspaceNameShould_Default_if_no_Name_Passed", func(t *testing.T) {
		GetByNameAndUserIdFunc = func(name string, userId string) ([]*Workspace, *restErrors.RestErr) {
			return []*Workspace{}, nil
		}
		CreateWorkspaceFunc = func(workspace *Workspace) *restErrors.RestErr {
			return nil
		}

		model, err := workspaceTestService.Create(&CreateWorkspaceRequestDto{Name: "default"}, "1")
		assert.Nil(t, err)
		assert.EqualValues(t, "default", model.Name)
	})

	t.Run("Create_Workspace_Should_throw_If_Name_Already_Exits_For_The_Same_User", func(t *testing.T) {
		GetByNameAndUserIdFunc = func(name string, userId string) ([]*Workspace, *restErrors.RestErr) {
			newWorkspace := new(Workspace)
			return []*Workspace{newWorkspace}, nil
		}

		model, err := workspaceTestService.Create(&CreateWorkspaceRequestDto{}, "1")
		assert.Nil(t, model)
		assert.EqualValues(t, http.StatusConflict, err.Status)
	})

	t.Run("WorkspaceNameShould_Throw_If_Create_User_in_Repo_Throws", func(t *testing.T) {
		GetByNameAndUserIdFunc = func(name string, userId string) ([]*Workspace, *restErrors.RestErr) {
			return []*Workspace{}, nil
		}
		CreateWorkspaceFunc = func(workspace *Workspace) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		model, err := workspaceTestService.Create(&CreateWorkspaceRequestDto{}, "1")
		assert.Nil(t, model)
		assert.EqualValues(t, http.StatusInternalServerError, err.Status)
	})
}

func TestService_Update(t *testing.T) {
	dto := new(UpdateWorkspaceRequestDto)
	dto.Name = "testName"
	t.Run("Update_Workspace_Should_Pass", func(t *testing.T) {
		UpdateWorkspaceFunc = func(workspace *Workspace) *restErrors.RestErr {
			return nil
		}
		GetByNameAndUserIdFunc = func(name string, userId string) ([]*Workspace, *restErrors.RestErr) {
			return []*Workspace{}, nil
		}

		model := new(Workspace)
		err := workspaceTestService.Update(dto, model)

		assert.Nil(t, err)
	})

	t.Run("Update_workspace_should_throw_if_repo_update_throws", func(t *testing.T) {
		model := new(Workspace)
		UpdateWorkspaceFunc = func(workspace *Workspace) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}
		GetByNameAndUserIdFunc = func(name string, userId string) ([]*Workspace, *restErrors.RestErr) {
			return []*Workspace{}, nil
		}

		err := workspaceTestService.Update(&UpdateWorkspaceRequestDto{}, model)
		assert.EqualValues(t, http.StatusInternalServerError, err.Status)
	})
	t.Run("Update_workspace_should_throw_if_repo_get_by_name_and_user_id_throws", func(t *testing.T) {
		model := new(Workspace)
		UpdateWorkspaceFunc = func(workspace *Workspace) *restErrors.RestErr {
			return nil
		}
		GetByNameAndUserIdFunc = func(name string, userId string) ([]*Workspace, *restErrors.RestErr) {
			return nil, restErrors.NewInternalServerError("something went wrong")
		}

		err := workspaceTestService.Update(&UpdateWorkspaceRequestDto{}, model)
		assert.EqualValues(t, http.StatusInternalServerError, err.Status)
	})

	t.Run("Update_Workspace_Should_Throw_if_name_to_update_already_exist_for_another_workspace_for_the_same_user", func(t *testing.T) {
		UpdateWorkspaceFunc = func(workspace *Workspace) *restErrors.RestErr {
			return nil
		}
		GetByNameAndUserIdFunc = func(name string, userId string) ([]*Workspace, *restErrors.RestErr) {
			newWorkspace := new(Workspace)
			newWorkspace.ID = uuid.NewString()
			newWorkspace.Name = dto.Name
			return []*Workspace{newWorkspace}, nil
		}

		model := new(Workspace)
		err := workspaceTestService.Update(dto, model)

		assert.EqualValues(t, "you have another workspace with the same name", err.Message)
	})

}

func TestService_GetById(t *testing.T) {
	t.Run("Get_work_space_by_id_should_pass", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*Workspace, *restErrors.RestErr) {
			return new(Workspace), nil
		}
		resp, err := workspaceTestService.GetById("1")
		assert.Nil(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("Get_work_space_by_id_should_Throw_if_repo_throws", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*Workspace, *restErrors.RestErr) {
			return nil, restErrors.NewNotFoundError("not found")
		}
		resp, err := workspaceTestService.GetById("1")
		assert.Nil(t, resp)
		assert.Error(t, err, "not found")
	})

}

func TestService_Delete(t *testing.T) {
	t.Run("Delete_workspace_should_pass", func(t *testing.T) {
		DeleteFunc = func(workspace *Workspace) *restErrors.RestErr {
			return nil
		}
		err := workspaceTestService.Delete(new(Workspace))
		assert.Nil(t, err)
	})

	t.Run("Delete_workspace_should_throw_if_repo_throws", func(t *testing.T) {
		DeleteFunc = func(workspace *Workspace) *restErrors.RestErr {
			return restErrors.NewNotFoundError("not found")
		}

		err := workspaceTestService.Delete(new(Workspace))
		assert.Error(t, err, "not found")
	})
}

func TestService_GetByUserId(t *testing.T) {
	t.Run("Get_by_user_id_should_pass", func(t *testing.T) {
		var list = make([]*Workspace, 0)
		workspace := new(Workspace)
		list = append(list, workspace)
		GetByUserIdFunc = func(userId string) ([]*Workspace, *restErrors.RestErr) {
			return list, nil
		}
		result, err := workspaceTestService.GetByUserId("1")
		assert.Nil(t, err)
		assert.EqualValues(t, list, result)
	})

	t.Run("Get_by_user_id_should_throw_if_repo_throws", func(t *testing.T) {
		GetByUserIdFunc = func(Id string) ([]*Workspace, *restErrors.RestErr) {
			return nil, restErrors.NewInternalServerError("something went wrong")
		}
		result, err := workspaceTestService.GetByUserId("1")
		assert.Nil(t, result)

		assert.Error(t, err, "something went wrong")
	})
}

func TestService_AddWorkspaceMember(t *testing.T) {
	t.Run("add_member_to_workspace_should_pass", func(t *testing.T) {
		addWorkspaceMemberFunc = func(workspace *Workspace, workspaceUser *workspaceuser.WorkspaceUser) *restErrors.RestErr {
			return nil
		}
		err := workspaceTestService.AddWorkspaceMember(new(Workspace), "1", roles.Writer)
		assert.Nil(t, err)
	})

	t.Run("add_member_to_workspace_should_throw_if_repo_throws", func(t *testing.T) {
		addWorkspaceMemberFunc = func(workspace *Workspace, workspaceUser *workspaceuser.WorkspaceUser) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		err := workspaceTestService.AddWorkspaceMember(new(Workspace), "1", roles.Writer)
		assert.Error(t, err, "something went wrong")
	})
}
func TestService_DeleteWorkspaceMember(t *testing.T) {
	t.Run("delete_workspace_member_should_pass", func(t *testing.T) {
		GetWorkspaceMemberByWorkspaceIdAndUserIdFunc = func(workspaceId string, userId string) (*workspaceuser.WorkspaceUser, *restErrors.RestErr) {
			return new(workspaceuser.WorkspaceUser), nil
		}
		DeleteWorkspaceMemberFunc = func(workspace *Workspace, workspaceUser *workspaceuser.WorkspaceUser) *restErrors.RestErr {
			return nil
		}
		err := workspaceTestService.DeleteWorkspaceMember(new(Workspace), "1")
		assert.Nil(t, err)
	})

	t.Run("delete_workspace_member_should_throw_if_workspace_member_doesnt_exist", func(t *testing.T) {
		GetWorkspaceMemberByWorkspaceIdAndUserIdFunc = func(workspaceId string, userId string) (*workspaceuser.WorkspaceUser, *restErrors.RestErr) {
			return nil, restErrors.NewNotFoundError("no such record")
		}

		err := workspaceTestService.DeleteWorkspaceMember(new(Workspace), "1")
		assert.Error(t, err, "no such record")
	})

	t.Run("delete_workspace_member_should_throw_if_repo_throws", func(t *testing.T) {
		GetWorkspaceMemberByWorkspaceIdAndUserIdFunc = func(workspaceId string, userId string) (*workspaceuser.WorkspaceUser, *restErrors.RestErr) {
			return new(workspaceuser.WorkspaceUser), nil
		}
		DeleteWorkspaceMemberFunc = func(workspace *Workspace, workspaceUser *workspaceuser.WorkspaceUser) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}
		err := workspaceTestService.DeleteWorkspaceMember(new(Workspace), "1")
		assert.Error(t, err, "something went wrong")
	})
}
func TestService_CountWorkspacesByUserId(t *testing.T) {
	t.Run("count_workspace_by_user_is_should_pass", func(t *testing.T) {
		CountByUserIdFunc = func(userId string) (int64, *restErrors.RestErr) {
			return 1, nil
		}

		result, err := workspaceTestService.CountByUserId("")
		assert.Nil(t, err)
		assert.EqualValues(t, 1, result)
	})

	t.Run("count_workspace_by_user_is_should_throw_if_repo_throws", func(t *testing.T) {
		CountByUserIdFunc = func(userId string) (int64, *restErrors.RestErr) {
			return 0, restErrors.NewInternalServerError("something went wrong")
		}

		result, err := workspaceTestService.CountByUserId("")
		assert.EqualValues(t, http.StatusInternalServerError, err.Status)
		assert.EqualValues(t, "something went wrong", err.Message)
		assert.EqualValues(t, 0, result)
	})

}

func TestService_UpdateWorkspaceUser(t *testing.T) {
	t.Run("update_workspace_user_should_pass", func(t *testing.T) {
		UpdateWorkspaceUserFunc = func(workspaceUser *workspaceuser.WorkspaceUser) *restErrors.RestErr {
			return nil
		}

		workspaceUser := new(workspaceuser.WorkspaceUser)
		dto := new(UpdateWorkspaceUserRequestDto)
		err := workspaceTestService.UpdateWorkspaceUser(workspaceUser, dto)
		assert.Nil(t, err)
	})

	t.Run("update_workspace_user_should_throw_if_repo_throws", func(t *testing.T) {
		UpdateWorkspaceUserFunc = func(workspaceUser *workspaceuser.WorkspaceUser) *restErrors.RestErr {
			return restErrors.NewInternalServerError("some thing went wrong")
		}

		workspaceUser := new(workspaceuser.WorkspaceUser)
		dto := new(UpdateWorkspaceUserRequestDto)
		err := workspaceTestService.UpdateWorkspaceUser(workspaceUser, dto)
		assert.NotNil(t, err)
		assert.EqualValues(t, "some thing went wrong", err.Message)
	})
}
