package workspace

import (
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/cloud-api/internal/workspaceuser"
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
	GetById                func(Id string) (*Workspace, *restErrors.RestErr)
	WithTransactionFunc    func(txHandle *gorm.DB) IRepository

	CreateWorkspaceUserFunc func(workspaceUser *workspaceuser.WorkspaceUser) *restErrors.RestErr
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
	return GetById(Id)
}

func (r workspaceRepositoryMock) WithTransaction(txHandle *gorm.DB) IRepository {
	return r
}

type workspaceuserRepositoryMock struct{}

func (workspaceuserRepositoryMock) Create(workspaceUser *workspaceuser.WorkspaceUser) *restErrors.RestErr {
	return CreateWorkspaceUserFunc(workspaceUser)
}
func (r workspaceuserRepositoryMock) WithTransaction(txHandle *gorm.DB) workspaceuser.IRepository {
	return r
}

func TestMain(m *testing.M) {
	workspaceRepo = &workspaceRepositoryMock{}
	workspaceUserRepo = &workspaceuserRepositoryMock{}

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
		CreateWorkspaceUserFunc = func(workspaceUser *workspaceuser.WorkspaceUser) *restErrors.RestErr {
			return nil
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
		CreateWorkspaceUserFunc = func(workspaceUser *workspaceuser.WorkspaceUser) *restErrors.RestErr {
			return nil
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

	t.Run("WorkspaceNameShould_Throw_If_Create_User_in_Workspace_User_Repo_Throws", func(t *testing.T) {
		GetByNameAndUserIdFunc = func(name string, userId string) (*Workspace, *restErrors.RestErr) {
			return nil, nil
		}
		CreateWorkspaceFunc = func(workspace *Workspace) *restErrors.RestErr {
			return nil
		}
		CreateWorkspaceUserFunc = func(workspaceUser *workspaceuser.WorkspaceUser) *restErrors.RestErr {
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
		GetById = func(userId string) (*Workspace, *restErrors.RestErr) {
			return new(Workspace), nil
		}

		UpdateWorkspaceFunc = func(workspace *Workspace) *restErrors.RestErr {
			return nil
		}

		model, err := workspaceTestService.Update(dto, "1")
		assert.Nil(t, err)
		assert.NotNil(t, model)
	})

	t.Run("WorkspaceNameShould_Throw_If_Create_User_in_Repo_Throws", func(t *testing.T) {
		GetById = func(userId string) (*Workspace, *restErrors.RestErr) {
			return new(Workspace), nil
		}
		UpdateWorkspaceFunc = func(workspace *Workspace) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		model, err := workspaceTestService.Update(&UpdateWorkspaceRequestDto{}, "1")
		assert.Nil(t, model)
		assert.EqualValues(t, http.StatusInternalServerError, err.Status)
	})

	t.Run("update_workspace_should_throw_if_workspace_does_not_exist", func(t *testing.T) {
		GetById = func(userId string) (*Workspace, *restErrors.RestErr) {
			return nil, restErrors.NewNotFoundError("not found")
		}

		model, err := workspaceTestService.Update(&UpdateWorkspaceRequestDto{}, "1")
		assert.Nil(t, model)
		assert.EqualValues(t, http.StatusNotFound, err.Status)
	})

}
