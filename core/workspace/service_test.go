package workspace

import (
	"github.com/google/uuid"
	"github.com/kotalco/core-api/core/workspaceuser"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/roles"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"net/http"
	"os"
	"testing"
)

/*
Workspace repo Mocks
*/
var (
	WithTransactionFunc                          func(txHandle *gorm.DB) IRepository
	workspaceTestService                         IService
	CreateWorkspaceFunc                          func(workspace *Workspace) restErrors.IRestErr
	UpdateWorkspaceFunc                          func(workspace *Workspace) restErrors.IRestErr
	GetByNameAndUserIdFunc                       func(name string, userId string) ([]*Workspace, restErrors.IRestErr)
	GetByIdFunc                                  func(Id string) (*Workspace, restErrors.IRestErr)
	DeleteFunc                                   func(workspace *Workspace) restErrors.IRestErr
	GetByUserIdFunc                              func(userId string) ([]*Workspace, restErrors.IRestErr)
	addWorkspaceMemberFunc                       func(workspace *Workspace, workspaceUser *workspaceuser.WorkspaceUser) restErrors.IRestErr
	DeleteWorkspaceMemberFunc                    func(workspace *Workspace, workspaceUser *workspaceuser.WorkspaceUser) restErrors.IRestErr
	GetWorkspaceMemberByWorkspaceIdAndUserIdFunc func(workspaceId string, userId string) (*workspaceuser.WorkspaceUser, restErrors.IRestErr)
	CountByUserIdFunc                            func(userId string) (int64, restErrors.IRestErr)
	UpdateWorkspaceUserFunc                      func(workspaceUser *workspaceuser.WorkspaceUser) restErrors.IRestErr
	GetByNamespaceFunc                           func(namespace string) (*Workspace, restErrors.IRestErr)
)

type workspaceRepositoryMock struct{}

func (r workspaceRepositoryMock) WithoutTransaction() IRepository {
	return r
}
func (r workspaceRepositoryMock) WithTransaction(txHandle *gorm.DB) IRepository {
	return r
}

func (r workspaceRepositoryMock) UpdateWorkspaceUser(workspaceUser *workspaceuser.WorkspaceUser) restErrors.IRestErr {
	return UpdateWorkspaceUserFunc(workspaceUser)
}
func (workspaceRepositoryMock) Create(workspace *Workspace) restErrors.IRestErr {
	return CreateWorkspaceFunc(workspace)
}

func (workspaceRepositoryMock) Update(workspace *Workspace) restErrors.IRestErr {
	return UpdateWorkspaceFunc(workspace)
}

func (workspaceRepositoryMock) GetByNameAndUserId(name string, userId string) ([]*Workspace, restErrors.IRestErr) {
	return GetByNameAndUserIdFunc(name, userId)
}
func (workspaceRepositoryMock) GetById(Id string) (*Workspace, restErrors.IRestErr) {
	return GetByIdFunc(Id)
}

func (workspaceRepositoryMock) Delete(workspace *Workspace) restErrors.IRestErr {
	return DeleteFunc(workspace)
}

func (workspaceRepositoryMock) GetByUserId(userId string) ([]*Workspace, restErrors.IRestErr) {
	return GetByUserIdFunc(userId)
}

func (workspaceRepositoryMock) AddWorkspaceMember(workspace *Workspace, workspaceUser *workspaceuser.WorkspaceUser) restErrors.IRestErr {
	return addWorkspaceMemberFunc(workspace, workspaceUser)
}

func (workspaceRepositoryMock) DeleteWorkspaceMember(workspace *Workspace, workspaceUser *workspaceuser.WorkspaceUser) restErrors.IRestErr {
	return DeleteWorkspaceMemberFunc(workspace, workspaceUser)
}
func (workspaceRepositoryMock) GetWorkspaceMemberByWorkspaceIdAndUserId(workspaceId string, userId string) (*workspaceuser.WorkspaceUser, restErrors.IRestErr) {
	return GetWorkspaceMemberByWorkspaceIdAndUserIdFunc(workspaceId, userId)
}
func (workspaceRepositoryMock) CountByUserId(userId string) (int64, restErrors.IRestErr) {
	return CountByUserIdFunc(userId)
}

func (workspaceRepositoryMock) GetByNamespace(namespace string) (*Workspace, restErrors.IRestErr) {
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
		GetByNameAndUserIdFunc = func(name string, userId string) ([]*Workspace, restErrors.IRestErr) {
			return []*Workspace{}, nil
		}
		CreateWorkspaceFunc = func(workspace *Workspace) restErrors.IRestErr {
			return nil
		}

		model, err := workspaceTestService.Create(dto, "1", uuid.NewString())
		assert.Nil(t, err)
		assert.NotNil(t, model)
		assert.EqualValues(t, roles.Admin, model.WorkspaceUsers[0].Role)
	})

	t.Run("WorkspaceNameShould_Default_if_no_Name_Passed", func(t *testing.T) {
		GetByNameAndUserIdFunc = func(name string, userId string) ([]*Workspace, restErrors.IRestErr) {
			return []*Workspace{}, nil
		}
		CreateWorkspaceFunc = func(workspace *Workspace) restErrors.IRestErr {
			return nil
		}

		model, err := workspaceTestService.Create(&CreateWorkspaceRequestDto{Name: "default"}, "1", uuid.NewString())
		assert.Nil(t, err)
		assert.EqualValues(t, "default", model.Name)
	})

	t.Run("Create_Workspace_Should_throw_If_Name_Already_Exits_For_The_Same_User", func(t *testing.T) {
		GetByNameAndUserIdFunc = func(name string, userId string) ([]*Workspace, restErrors.IRestErr) {
			newWorkspace := new(Workspace)
			return []*Workspace{newWorkspace}, nil
		}

		model, err := workspaceTestService.Create(&CreateWorkspaceRequestDto{}, "1", uuid.NewString())
		assert.Nil(t, model)
		assert.EqualValues(t, http.StatusConflict, err.StatusCode())
	})

	t.Run("WorkspaceNameShould_Throw_If_Create_User_in_Repo_Throws", func(t *testing.T) {
		GetByNameAndUserIdFunc = func(name string, userId string) ([]*Workspace, restErrors.IRestErr) {
			return []*Workspace{}, nil
		}
		CreateWorkspaceFunc = func(workspace *Workspace) restErrors.IRestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		model, err := workspaceTestService.Create(&CreateWorkspaceRequestDto{}, "1", uuid.NewString())
		assert.Nil(t, model)
		assert.EqualValues(t, http.StatusInternalServerError, err.StatusCode())
	})
}

func TestService_Update(t *testing.T) {
	dto := new(UpdateWorkspaceRequestDto)
	dto.Name = "testName"
	t.Run("Update_Workspace_Should_Pass", func(t *testing.T) {
		UpdateWorkspaceFunc = func(workspace *Workspace) restErrors.IRestErr {
			return nil
		}
		GetByNameAndUserIdFunc = func(name string, userId string) ([]*Workspace, restErrors.IRestErr) {
			return []*Workspace{}, nil
		}

		model := new(Workspace)
		err := workspaceTestService.Update(dto, model)

		assert.Nil(t, err)
	})

	t.Run("Update_workspace_should_throw_if_repo_update_throws", func(t *testing.T) {
		model := new(Workspace)
		UpdateWorkspaceFunc = func(workspace *Workspace) restErrors.IRestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}
		GetByNameAndUserIdFunc = func(name string, userId string) ([]*Workspace, restErrors.IRestErr) {
			return []*Workspace{}, nil
		}

		err := workspaceTestService.Update(&UpdateWorkspaceRequestDto{}, model)
		assert.EqualValues(t, http.StatusInternalServerError, err.StatusCode())
	})
	t.Run("Update_workspace_should_throw_if_repo_get_by_name_and_user_id_throws", func(t *testing.T) {
		model := new(Workspace)
		UpdateWorkspaceFunc = func(workspace *Workspace) restErrors.IRestErr {
			return nil
		}
		GetByNameAndUserIdFunc = func(name string, userId string) ([]*Workspace, restErrors.IRestErr) {
			return nil, restErrors.NewInternalServerError("something went wrong")
		}

		err := workspaceTestService.Update(&UpdateWorkspaceRequestDto{}, model)
		assert.EqualValues(t, http.StatusInternalServerError, err.StatusCode())
	})

	t.Run("Update_Workspace_Should_Throw_if_name_to_update_already_exist_for_another_workspace_for_the_same_user", func(t *testing.T) {
		UpdateWorkspaceFunc = func(workspace *Workspace) restErrors.IRestErr {
			return nil
		}
		GetByNameAndUserIdFunc = func(name string, userId string) ([]*Workspace, restErrors.IRestErr) {
			newWorkspace := new(Workspace)
			newWorkspace.ID = uuid.NewString()
			newWorkspace.Name = dto.Name
			return []*Workspace{newWorkspace}, nil
		}

		model := new(Workspace)
		err := workspaceTestService.Update(dto, model)

		assert.EqualValues(t, "you have another workspace with the same name", err.Error())
	})

}

func TestService_GetById(t *testing.T) {
	t.Run("Get_work_space_by_id_should_pass", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*Workspace, restErrors.IRestErr) {
			return new(Workspace), nil
		}
		resp, err := workspaceTestService.GetById("1")
		assert.Nil(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("Get_work_space_by_id_should_Throw_if_repo_throws", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*Workspace, restErrors.IRestErr) {
			return nil, restErrors.NewNotFoundError("not found")
		}
		resp, err := workspaceTestService.GetById("1")
		assert.Nil(t, resp)
		assert.Error(t, err, "not found")
	})

}

func TestService_Delete(t *testing.T) {
	t.Run("Delete_workspace_should_pass", func(t *testing.T) {
		DeleteFunc = func(workspace *Workspace) restErrors.IRestErr {
			return nil
		}
		err := workspaceTestService.Delete(new(Workspace))
		assert.Nil(t, err)
	})

	t.Run("Delete_workspace_should_throw_if_repo_throws", func(t *testing.T) {
		DeleteFunc = func(workspace *Workspace) restErrors.IRestErr {
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
		GetByUserIdFunc = func(userId string) ([]*Workspace, restErrors.IRestErr) {
			return list, nil
		}
		result, err := workspaceTestService.GetByUserId("1")
		assert.Nil(t, err)
		assert.EqualValues(t, list, result)
	})

	t.Run("Get_by_user_id_should_throw_if_repo_throws", func(t *testing.T) {
		GetByUserIdFunc = func(Id string) ([]*Workspace, restErrors.IRestErr) {
			return nil, restErrors.NewInternalServerError("something went wrong")
		}
		result, err := workspaceTestService.GetByUserId("1")
		assert.Nil(t, result)

		assert.Error(t, err, "something went wrong")
	})
}

func TestService_AddWorkspaceMember(t *testing.T) {
	t.Run("add_member_to_workspace_should_pass", func(t *testing.T) {
		addWorkspaceMemberFunc = func(workspace *Workspace, workspaceUser *workspaceuser.WorkspaceUser) restErrors.IRestErr {
			return nil
		}
		err := workspaceTestService.AddWorkspaceMember(new(Workspace), "1", roles.Writer)
		assert.Nil(t, err)
	})

	t.Run("add_member_to_workspace_should_throw_if_repo_throws", func(t *testing.T) {
		addWorkspaceMemberFunc = func(workspace *Workspace, workspaceUser *workspaceuser.WorkspaceUser) restErrors.IRestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		err := workspaceTestService.AddWorkspaceMember(new(Workspace), "1", roles.Writer)
		assert.Error(t, err, "something went wrong")
	})
}
func TestService_DeleteWorkspaceMember(t *testing.T) {
	t.Run("delete_workspace_member_should_pass", func(t *testing.T) {
		GetWorkspaceMemberByWorkspaceIdAndUserIdFunc = func(workspaceId string, userId string) (*workspaceuser.WorkspaceUser, restErrors.IRestErr) {
			return new(workspaceuser.WorkspaceUser), nil
		}
		DeleteWorkspaceMemberFunc = func(workspace *Workspace, workspaceUser *workspaceuser.WorkspaceUser) restErrors.IRestErr {
			return nil
		}
		err := workspaceTestService.DeleteWorkspaceMember(new(Workspace), "1")
		assert.Nil(t, err)
	})

	t.Run("delete_workspace_member_should_throw_if_workspace_member_doesnt_exist", func(t *testing.T) {
		GetWorkspaceMemberByWorkspaceIdAndUserIdFunc = func(workspaceId string, userId string) (*workspaceuser.WorkspaceUser, restErrors.IRestErr) {
			return nil, restErrors.NewNotFoundError("no such record")
		}

		err := workspaceTestService.DeleteWorkspaceMember(new(Workspace), "1")
		assert.Error(t, err, "no such record")
	})

	t.Run("delete_workspace_member_should_throw_if_repo_throws", func(t *testing.T) {
		GetWorkspaceMemberByWorkspaceIdAndUserIdFunc = func(workspaceId string, userId string) (*workspaceuser.WorkspaceUser, restErrors.IRestErr) {
			return new(workspaceuser.WorkspaceUser), nil
		}
		DeleteWorkspaceMemberFunc = func(workspace *Workspace, workspaceUser *workspaceuser.WorkspaceUser) restErrors.IRestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}
		err := workspaceTestService.DeleteWorkspaceMember(new(Workspace), "1")
		assert.Error(t, err, "something went wrong")
	})
}
func TestService_CountWorkspacesByUserId(t *testing.T) {
	t.Run("count_workspace_by_user_is_should_pass", func(t *testing.T) {
		CountByUserIdFunc = func(userId string) (int64, restErrors.IRestErr) {
			return 1, nil
		}

		result, err := workspaceTestService.CountByUserId("")
		assert.Nil(t, err)
		assert.EqualValues(t, 1, result)
	})

	t.Run("count_workspace_by_user_is_should_throw_if_repo_throws", func(t *testing.T) {
		CountByUserIdFunc = func(userId string) (int64, restErrors.IRestErr) {
			return 0, restErrors.NewInternalServerError("something went wrong")
		}

		result, err := workspaceTestService.CountByUserId("")
		assert.EqualValues(t, http.StatusInternalServerError, err.StatusCode())
		assert.EqualValues(t, "something went wrong", err.Error())
		assert.EqualValues(t, 0, result)
	})

}

func TestService_UpdateWorkspaceUser(t *testing.T) {
	t.Run("update_workspace_user_should_pass", func(t *testing.T) {
		UpdateWorkspaceUserFunc = func(workspaceUser *workspaceuser.WorkspaceUser) restErrors.IRestErr {
			return nil
		}

		workspaceUser := new(workspaceuser.WorkspaceUser)
		dto := new(UpdateWorkspaceUserRequestDto)
		err := workspaceTestService.UpdateWorkspaceUser(workspaceUser, dto)
		assert.Nil(t, err)
	})

	t.Run("update_workspace_user_should_throw_if_repo_throws", func(t *testing.T) {
		UpdateWorkspaceUserFunc = func(workspaceUser *workspaceuser.WorkspaceUser) restErrors.IRestErr {
			return restErrors.NewInternalServerError("some thing went wrong")
		}

		workspaceUser := new(workspaceuser.WorkspaceUser)
		dto := new(UpdateWorkspaceUserRequestDto)
		err := workspaceTestService.UpdateWorkspaceUser(workspaceUser, dto)
		assert.NotNil(t, err)
		assert.EqualValues(t, "some thing went wrong", err.Error())
	})
}
