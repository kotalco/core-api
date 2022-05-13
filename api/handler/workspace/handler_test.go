package workspace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/cloud-api/internal/workspace"
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
	ListWorkspacesFunc       func(userId string) ([]*workspace.WorkspaceResponseDto, *restErrors.RestErr)
	DeleteWorkspace          func(id string) *restErrors.RestErr
	WorkspaceWithTransaction func(txHandle *gorm.DB) workspace.IService
)

type workspaceServiceMock struct{}

func (workspaceServiceMock) Create(dto *workspace.CreateWorkspaceRequestDto, userId string) (*workspace.Workspace, *restErrors.RestErr) {
	return CreateWorkspaceFunc(dto, userId)
}

func (workspaceServiceMock) List(userId string) ([]*workspace.WorkspaceResponseDto, *restErrors.RestErr) {
	return ListWorkspacesFunc(userId)
}

func (workspaceServiceMock) Delete(id string) *restErrors.RestErr {
	return DeleteWorkspace(id)
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
	//var invalidDto = map[string]string{
	//	"email": "Invalid_Name",
	//}
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

		fmt.Println(string(body), resp.StatusCode)
		var result map[string]workspace.WorkspaceResponseDto
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusCreated, resp.StatusCode)
		assert.EqualValues(t, "testName", result["data"].Name)
	})
}
