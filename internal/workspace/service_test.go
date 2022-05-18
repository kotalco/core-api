package workspace

import (
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"net/http"
	"os"
	"testing"
)

var (
	workspaceTestService   IService
	CreateWorkspaceFunc    func(workspace *Workspace) *restErrors.RestErr
	UpdateWorkspaceFunc    func(workspace *Workspace) *restErrors.RestErr
	GetByNameAndUserIdFunc func(name string, userId string) (*Workspace, *restErrors.RestErr)
	GetByIdFunc            func(Id string) (*Workspace, *restErrors.RestErr)
	DeleteFunc             func(workspace *Workspace) *restErrors.RestErr
	GetByUserIdFunc        func(userId string) ([]Workspace, *restErrors.RestErr)
	WithTransactionFunc    func(txHandle *gorm.DB) IRepository
)

type workspaceRepositoryMock struct{}

func (workspaceRepositoryMock) Create(workspace *Workspace) *restErrors.RestErr {
	return CreateWorkspaceFunc(workspace)
}

func (workspaceRepositoryMock) Update(workspace *Workspace) *restErrors.RestErr {
	return UpdateWorkspaceFunc(workspace)
}

func (workspaceRepositoryMock) GetByNameAndUserId(name string, userId string) (*Workspace, *restErrors.RestErr) {
	return GetByNameAndUserIdFunc(name, userId)
}
func (workspaceRepositoryMock) GetById(Id string) (*Workspace, *restErrors.RestErr) {
	return GetByIdFunc(Id)
}

func (workspaceRepositoryMock) Delete(workspace *Workspace) *restErrors.RestErr {
	return DeleteFunc(workspace)
}

func (workspaceRepositoryMock) GetByUserId(userId string) ([]Workspace, *restErrors.RestErr) {
	return GetByUserIdFunc(userId)
}

func (r workspaceRepositoryMock) WithTransaction(txHandle *gorm.DB) IRepository {
	return r
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
		GetByNameAndUserIdFunc = func(name string, userId string) (*Workspace, *restErrors.RestErr) {
			return nil, nil
		}
		CreateWorkspaceFunc = func(workspace *Workspace) *restErrors.RestErr {
			return nil
		}

		model, err := workspaceTestService.Create(dto, "1")
		assert.Nil(t, err)
		assert.NotNil(t, model)
	})

	t.Run("WorkspaceNameShould_Default_if_no_Name_Passed", func(t *testing.T) {
		GetByNameAndUserIdFunc = func(name string, userId string) (*Workspace, *restErrors.RestErr) {
			return nil, nil
		}
		CreateWorkspaceFunc = func(workspace *Workspace) *restErrors.RestErr {
			return nil
		}

		model, err := workspaceTestService.Create(&CreateWorkspaceRequestDto{Name: "default"}, "1")
		assert.Nil(t, err)
		assert.EqualValues(t, "default", model.Name)
	})

	t.Run("Create_Workspace_Should_throw_If_Name_Already_Exits_For_The_Same_User", func(t *testing.T) {
		GetByNameAndUserIdFunc = func(name string, userId string) (*Workspace, *restErrors.RestErr) {
			return new(Workspace), nil
		}

		model, err := workspaceTestService.Create(&CreateWorkspaceRequestDto{}, "1")
		assert.Nil(t, model)
		assert.EqualValues(t, http.StatusConflict, err.Status)
	})

	t.Run("WorkspaceNameShould_Throw_If_Create_User_in_Repo_Throws", func(t *testing.T) {
		GetByNameAndUserIdFunc = func(name string, userId string) (*Workspace, *restErrors.RestErr) {
			return nil, nil
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

		model := new(Workspace)
		err := workspaceTestService.Update(dto, model)

		assert.Nil(t, err)
	})

	t.Run("Update_workspace_should_throw_if_repo_update_throws", func(t *testing.T) {
		model := new(Workspace)
		UpdateWorkspaceFunc = func(workspace *Workspace) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		err := workspaceTestService.Update(&UpdateWorkspaceRequestDto{}, model)
		assert.EqualValues(t, http.StatusInternalServerError, err.Status)
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
