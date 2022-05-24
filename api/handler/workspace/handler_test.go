package workspace

import (
	"bytes"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/cloud-api/internal/workspace"
	"github.com/kotalco/cloud-api/internal/workspaceuser"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	"github.com/kotalco/cloud-api/pkg/token"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

/*
Workspace service Mocks
*/
var (
	CreateWorkspaceFunc      func(dto *workspace.CreateWorkspaceRequestDto, userId string) (*workspace.Workspace, *restErrors.RestErr)
	UpdateWorkspaceFunc      func(dto *workspace.UpdateWorkspaceRequestDto, workspace *workspace.Workspace) *restErrors.RestErr
	GetWorkspaceByIdFunc     func(Id string) (*workspace.Workspace, *restErrors.RestErr)
	DeleteWorkspaceFunc      func(workspace *workspace.Workspace) *restErrors.RestErr
	GetWorkspaceByUserId     func(userId string) ([]*workspace.Workspace, *restErrors.RestErr)
	WorkspaceWithTransaction func(txHandle *gorm.DB) workspace.IService
)

type workspaceServiceMock struct{}

func (workspaceServiceMock) Create(dto *workspace.CreateWorkspaceRequestDto, userId string) (*workspace.Workspace, *restErrors.RestErr) {
	return CreateWorkspaceFunc(dto, userId)
}
func (workspaceServiceMock) Update(dto *workspace.UpdateWorkspaceRequestDto, workspace *workspace.Workspace) *restErrors.RestErr {
	return UpdateWorkspaceFunc(dto, workspace)
}
func (workspaceServiceMock) GetById(workspaceId string) (*workspace.Workspace, *restErrors.RestErr) {
	return GetWorkspaceByIdFunc(workspaceId)
}
func (workspaceServiceMock) Delete(workspace *workspace.Workspace) *restErrors.RestErr {
	return DeleteWorkspaceFunc(workspace)
}

func (workspaceServiceMock) GetByUserId(workspaceId string) ([]*workspace.Workspace, *restErrors.RestErr) {
	return GetWorkspaceByUserId(workspaceId)
}

func (wService workspaceServiceMock) WithTransaction(txHandle *gorm.DB) workspace.IService {
	return wService
}

/*
Namespace service Mocks
*/

var (
	CreateNamespaceFunc func(name string) *restErrors.RestErr
	GetNamespaceFunc    func(name string) (*corev1.Namespace, *restErrors.RestErr)
	DeleteNamespaceFunc func(name string) *restErrors.RestErr
)

type namespaceServiceMock struct{}

func (namespaceServiceMock) Create(name string) *restErrors.RestErr {
	return CreateNamespaceFunc(name)
}

func (namespaceServiceMock) Get(name string) (*corev1.Namespace, *restErrors.RestErr) {
	return GetNamespaceFunc(name)
}

func (namespaceServiceMock) Delete(name string) *restErrors.RestErr {
	return DeleteNamespaceFunc(name)
}

func newFiberCtx(dto interface{}, method func(c *fiber.Ctx) error, locals map[string]interface{}) ([]byte, *http.Response) {
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		for key, element := range locals {
			c.Locals(key, element)
		}
		return method(c)
	})

	marshaledDto, err := json.Marshal(dto)
	if err != nil {
		panic(err.Error())
	}

	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(marshaledDto))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	return body, resp
}

func TestMain(m *testing.M) {
	workspaceService = &workspaceServiceMock{}
	namespaceService = &namespaceServiceMock{}
	sqlclient.OpenDBConnection()

	code := m.Run()
	os.Exit(code)
}

func TestCreateWorkspace(t *testing.T) {
	userDetails := new(token.UserDetails)
	userDetails.ID = "test@test.com"
	var locals = map[string]interface{}{}
	locals["user"] = *userDetails

	var validDto = map[string]string{
		"name": "testnamespace",
	}
	var invalidDto = map[string]string{
		"name": "",
	}
	t.Run("create_workspace_should_pass", func(t *testing.T) {
		CreateNamespaceFunc = func(name string) *restErrors.RestErr {
			return nil
		}

		CreateWorkspaceFunc = func(dto *workspace.CreateWorkspaceRequestDto, userId string) (*workspace.Workspace, *restErrors.RestErr) {
			model := new(workspace.Workspace)
			model.ID = "1"
			model.Name = "testName"
			return model, nil
		}

		body, resp := newFiberCtx(validDto, Create, locals)

		var result map[string]workspace.WorkspaceResponseDto
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusCreated, resp.StatusCode)
		assert.EqualValues(t, "testName", result["data"].Name)

	})

	t.Run("create_workspace_should_throw_validation_error", func(t *testing.T) {
		body, resp := newFiberCtx(invalidDto, Create, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}
		var fields = map[string]string{}
		fields["name"] = "name should be greater than 1 char and less than 100 char"
		badReqErr := restErrors.NewValidationError(fields)

		assert.EqualValues(t, badReqErr.Status, resp.StatusCode)
		assert.Equal(t, *badReqErr, result)
	})

	t.Run("Create_workspace_Should_Throw_Invalid_Request_Error", func(t *testing.T) {
		body, resp := newFiberCtx("", Create, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "invalid request body", result.Message)
	})

	t.Run("Create_Workspace_Should_Throw_If_workspace_Service_Throw", func(t *testing.T) {
		CreateWorkspaceFunc = func(dto *workspace.CreateWorkspaceRequestDto, userId string) (*workspace.Workspace, *restErrors.RestErr) {
			return nil, restErrors.NewBadRequestError("workspace service error")

		}

		body, resp := newFiberCtx(validDto, Create, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "workspace service error", result.Message)
	})

	t.Run("create_workspace_should_throw_if_can't_create_namespace", func(t *testing.T) {

		CreateWorkspaceFunc = func(dto *workspace.CreateWorkspaceRequestDto, userId string) (*workspace.Workspace, *restErrors.RestErr) {
			model := new(workspace.Workspace)
			model.ID = "1"
			model.Name = "testName"
			return model, nil
		}
		CreateNamespaceFunc = func(name string) *restErrors.RestErr {
			return restErrors.NewInternalServerError("can't create namespace")
		}

		body, resp := newFiberCtx(validDto, Create, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)
		assert.EqualValues(t, "can't create namespace", result.Message)
	})
}

func TestUpdateWorkspace(t *testing.T) {
	userDetails := new(token.UserDetails)
	userDetails.ID = "test@test.com"
	var locals = map[string]interface{}{}

	workspaceModelLocals := new(workspace.Workspace)
	workspaceUserModelLocals := new(workspaceuser.WorkspaceUser)
	workspaceUserModelLocals.UserId = userDetails.ID
	workspaceModelLocals.WorkspaceUsers = []workspaceuser.WorkspaceUser{*workspaceUserModelLocals}

	locals["user"] = *userDetails
	locals["workspace"] = *workspaceModelLocals

	var validDto = map[string]string{
		"name": "testnamespace",
	}
	var invalidDto = map[string]string{
		"name": "",
	}
	t.Run("update_workspace_should_pass", func(t *testing.T) {

		newWorkspace := new(workspace.Workspace)
		newWorkspace.UserId = userDetails.ID

		newWorkspaceUser := new(workspaceuser.WorkspaceUser)
		newWorkspaceUser.UserId = userDetails.ID

		newWorkspace.WorkspaceUsers = []workspaceuser.WorkspaceUser{*newWorkspaceUser}

		UpdateWorkspaceFunc = func(dto *workspace.UpdateWorkspaceRequestDto, workspace *workspace.Workspace) *restErrors.RestErr {
			return nil
		}

		body, resp := newFiberCtx(validDto, Update, locals)
		var result map[string]workspace.WorkspaceResponseDto
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
		assert.NotNil(t, newWorkspace.Name, result["data"].Name)

	})

	t.Run("update_workspace_should_throw_validation_error", func(t *testing.T) {
		body, resp := newFiberCtx(invalidDto, Update, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}
		var fields = map[string]string{}
		fields["name"] = "name should be greater than 1 char and less than 100 char"
		badReqErr := restErrors.NewValidationError(fields)

		assert.EqualValues(t, badReqErr.Status, resp.StatusCode)
		assert.Equal(t, *badReqErr, result)
	})

	t.Run("Update_workspace_Should_Throw_Invalid_Request_Error", func(t *testing.T) {
		body, resp := newFiberCtx("", Update, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "invalid request body", result.Message)
	})

	t.Run("Update_Workspace_Should_Throw_If_workspace_Service_Throw", func(t *testing.T) {
		newWorkspace := new(workspace.Workspace)
		newWorkspace.UserId = userDetails.ID

		newWorkspaceUser := new(workspaceuser.WorkspaceUser)
		newWorkspaceUser.UserId = userDetails.ID

		newWorkspace.WorkspaceUsers = []workspaceuser.WorkspaceUser{*newWorkspaceUser}

		GetWorkspaceByIdFunc = func(Id string) (*workspace.Workspace, *restErrors.RestErr) {
			return newWorkspace, nil
		}

		UpdateWorkspaceFunc = func(dto *workspace.UpdateWorkspaceRequestDto, workspace *workspace.Workspace) *restErrors.RestErr {
			return restErrors.NewInternalServerError("workspace service error")
		}

		body, resp := newFiberCtx(validDto, Update, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)
		assert.EqualValues(t, "workspace service error", result.Message)
	})
}

func TestDeleteWorkspace(t *testing.T) {
	userDetails := new(token.UserDetails)
	userDetails.ID = "11"

	workspaceModelLocals := new(workspace.Workspace)
	workspaceUserModelLocals := new(workspaceuser.WorkspaceUser)
	workspaceUserModelLocals.UserId = userDetails.ID
	workspaceModelLocals.WorkspaceUsers = []workspaceuser.WorkspaceUser{*workspaceUserModelLocals}

	var locals = map[string]interface{}{}
	locals["user"] = *userDetails
	locals["workspace"] = *workspaceModelLocals

	t.Run("Delete_Workspace_should_pass", func(t *testing.T) {

		DeleteWorkspaceFunc = func(workspace *workspace.Workspace) *restErrors.RestErr {
			return nil
		}

		DeleteNamespaceFunc = func(name string) *restErrors.RestErr {
			return nil
		}

		_, resp := newFiberCtx("", Delete, locals)
		assert.EqualValues(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("Delete_Workspace_should_throw_if_workspace_repo_throws", func(t *testing.T) {
		DeleteWorkspaceFunc = func(workspace *workspace.Workspace) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		body, _ := newFiberCtx("", Delete, locals)

		var restErr restErrors.RestErr
		err := json.Unmarshal(body, &restErr)
		if err != nil {
			panic(err)
		}
		assert.EqualValues(t, http.StatusInternalServerError, restErr.Status)
		assert.Error(t, restErr, "something went wrong")
	})

	t.Run("Delete_Workspace_should_throw_if_namespace_service_throws", func(t *testing.T) {
		DeleteWorkspaceFunc = func(workspace *workspace.Workspace) *restErrors.RestErr {
			return nil
		}

		DeleteNamespaceFunc = func(name string) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		body, _ := newFiberCtx("", Delete, locals)

		var restErr restErrors.RestErr
		err := json.Unmarshal(body, &restErr)
		if err != nil {
			panic(err)
		}
		assert.EqualValues(t, http.StatusInternalServerError, restErr.Status)
		assert.Error(t, restErr, "something went wrong")
	})

}

func TestGetWorkspaceByUserId(t *testing.T) {
	userDetails := new(token.UserDetails)
	userDetails.ID = "11"
	var locals = map[string]interface{}{}
	locals["user"] = *userDetails

	t.Run("Get_workspace_by_user_is_should_pass", func(t *testing.T) {
		GetWorkspaceByUserId = func(userId string) ([]*workspace.Workspace, *restErrors.RestErr) {
			var list = make([]*workspace.Workspace, 0)
			record := new(workspace.Workspace)
			list = append(list, record)
			return list, nil
		}

		result, resp := newFiberCtx("", GetByUserId, locals)
		var workspaceList map[string][]workspace.Workspace
		err := json.Unmarshal(result, &workspaceList)
		if err != nil {
			panic(err)
		}
		assert.NotNil(t, workspaceList)
		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Get_workspace_by_user_is_should_throw_if_workspace_service_throws", func(t *testing.T) {
		GetWorkspaceByUserId = func(userId string) ([]*workspace.Workspace, *restErrors.RestErr) {
			return nil, restErrors.NewInternalServerError("something went wrong")
		}
		result, _ := newFiberCtx("", GetByUserId, locals)
		var restErr restErrors.RestErr
		err := json.Unmarshal(result, &restErr)
		if err != nil {
			panic(err)
		}
		assert.EqualValues(t, http.StatusInternalServerError, restErr.Status)
	})

}
