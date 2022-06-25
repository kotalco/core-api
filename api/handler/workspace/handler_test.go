package workspace

import (
	"bytes"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/api/pkg/shared"
	"github.com/kotalco/cloud-api/internal/user"
	"github.com/kotalco/cloud-api/internal/workspace"
	"github.com/kotalco/cloud-api/internal/workspaceuser"
	"github.com/kotalco/cloud-api/pkg/roles"
	"github.com/kotalco/cloud-api/pkg/sendgrid"
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
User service Mocks
*/
var (
	UserWithTransactionFunc  func(txHandle *gorm.DB) user.IService
	SignUpFunc               func(dto *user.SignUpRequestDto) (*user.User, *restErrors.RestErr)
	SignInFunc               func(dto *user.SignInRequestDto) (*user.UserSessionResponseDto, *restErrors.RestErr)
	VerifyTOTPFunc           func(model *user.User, totp string) (*user.UserSessionResponseDto, *restErrors.RestErr)
	GetByEmailFunc           func(email string) (*user.User, *restErrors.RestErr)
	GetByIdFunc              func(Id string) (*user.User, *restErrors.RestErr)
	VerifyEmailFunc          func(model *user.User) *restErrors.RestErr
	ResetPasswordFunc        func(model *user.User, password string) *restErrors.RestErr
	ChangePasswordFunc       func(model *user.User, dto *user.ChangePasswordRequestDto) *restErrors.RestErr
	ChangeEmailFunc          func(model *user.User, dto *user.ChangeEmailRequestDto) *restErrors.RestErr
	CreateTOTPFunc           func(model *user.User, dto *user.CreateTOTPRequestDto) (bytes.Buffer, *restErrors.RestErr)
	EnableTwoFactorAuthFunc  func(model *user.User, totp string) (*user.User, *restErrors.RestErr)
	DisableTwoFactorAuthFunc func(model *user.User, dto *user.DisableTOTPRequestDto) *restErrors.RestErr
	FindWhereIdInSliceFunc   func(ids []string) ([]*user.User, *restErrors.RestErr)
)

type userServiceMock struct{}

func (uService userServiceMock) WithTransaction(txHandle *gorm.DB) user.IService {
	return uService
}

func (userServiceMock) SignUp(dto *user.SignUpRequestDto) (*user.User, *restErrors.RestErr) {
	return SignUpFunc(dto)
}

func (userServiceMock) SignIn(dto *user.SignInRequestDto) (*user.UserSessionResponseDto, *restErrors.RestErr) {
	return SignInFunc(dto)
}

func (userServiceMock) GetByEmail(email string) (*user.User, *restErrors.RestErr) {
	return GetByEmailFunc(email)
}

func (userServiceMock) GetById(Id string) (*user.User, *restErrors.RestErr) {
	return GetByIdFunc(Id)
}

func (userServiceMock) VerifyEmail(model *user.User) *restErrors.RestErr {
	return VerifyEmailFunc(model)
}

func (userServiceMock) ResetPassword(model *user.User, password string) *restErrors.RestErr {
	return ResetPasswordFunc(model, password)
}

func (userServiceMock) ChangePassword(model *user.User, dto *user.ChangePasswordRequestDto) *restErrors.RestErr {
	return ChangePasswordFunc(model, dto)
}

func (userServiceMock) ChangeEmail(model *user.User, dto *user.ChangeEmailRequestDto) *restErrors.RestErr {
	return ChangeEmailFunc(model, dto)
}

func (userServiceMock) CreateTOTP(model *user.User, dto *user.CreateTOTPRequestDto) (bytes.Buffer, *restErrors.RestErr) {
	return CreateTOTPFunc(model, dto)
}

func (userServiceMock) EnableTwoFactorAuth(model *user.User, totp string) (*user.User, *restErrors.RestErr) {
	return EnableTwoFactorAuthFunc(model, totp)
}

func (userServiceMock) VerifyTOTP(model *user.User, totp string) (*user.UserSessionResponseDto, *restErrors.RestErr) {
	return VerifyTOTPFunc(model, totp)
}

func (userServiceMock) DisableTwoFactorAuth(model *user.User, dto *user.DisableTOTPRequestDto) *restErrors.RestErr {
	return DisableTwoFactorAuthFunc(model, dto)
}

func (userServiceMock) FindWhereIdInSlice(ids []string) ([]*user.User, *restErrors.RestErr) {
	return FindWhereIdInSliceFunc(ids)
}

/*
Workspace service Mocks
*/
var (
	WorkspaceWithTransaction   func(txHandle *gorm.DB) workspace.IService
	CreateWorkspaceFunc        func(dto *workspace.CreateWorkspaceRequestDto, userId string) (*workspace.Workspace, *restErrors.RestErr)
	UpdateWorkspaceFunc        func(dto *workspace.UpdateWorkspaceRequestDto, workspace *workspace.Workspace) *restErrors.RestErr
	GetWorkspaceByIdFunc       func(Id string) (*workspace.Workspace, *restErrors.RestErr)
	DeleteWorkspaceFunc        func(workspace *workspace.Workspace) *restErrors.RestErr
	GetWorkspaceByUserIdFunc   func(userId string) ([]*workspace.Workspace, *restErrors.RestErr)
	AddWorkspaceMemberFunc     func(workspace *workspace.Workspace, memberId string, role string) *restErrors.RestErr
	DeleteWorkspaceMemberFunc  func(workspace *workspace.Workspace, memberId string) *restErrors.RestErr
	CountWorkspaceByUserIdFunc func(userId string) (int64, *restErrors.RestErr)
)

type workspaceServiceMock struct{}

func (wService workspaceServiceMock) WithTransaction(txHandle *gorm.DB) workspace.IService {
	return wService
}

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
	return GetWorkspaceByUserIdFunc(workspaceId)
}

func (workspaceServiceMock) AddWorkspaceMember(workspace *workspace.Workspace, memberId string, role string) *restErrors.RestErr {
	return AddWorkspaceMemberFunc(workspace, memberId, role)
}

func (workspaceServiceMock) DeleteWorkspaceMember(workspace *workspace.Workspace, memberId string) *restErrors.RestErr {
	return DeleteWorkspaceMemberFunc(workspace, memberId)
}
func (workspaceServiceMock) CountByUserId(userId string) (int64, *restErrors.RestErr) {
	return CountWorkspaceByUserIdFunc(userId)
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

//Mail Service mocks
var (
	SignUpMailFunc              func(dto *sendgrid.MailRequestDto) *restErrors.RestErr
	ResendEmailVerificationFunc func(dto *sendgrid.MailRequestDto) *restErrors.RestErr
	ForgetPasswordMailFunc      func(dto *sendgrid.MailRequestDto) *restErrors.RestErr
	WorkspaceInvitationFunc     func(dto *sendgrid.WorkspaceInvitationMailRequestDto) *restErrors.RestErr
)

type mailServiceMock struct{}

func (mailServiceMock) SignUp(dto *sendgrid.MailRequestDto) *restErrors.RestErr {
	return SignUpMailFunc(dto)
}
func (mailServiceMock) ResendEmailVerification(dto *sendgrid.MailRequestDto) *restErrors.RestErr {
	return ResendEmailVerificationFunc(dto)
}
func (mailServiceMock) ForgetPassword(dto *sendgrid.MailRequestDto) *restErrors.RestErr {
	return ForgetPasswordMailFunc(dto)
}
func (mailServiceMock) WorkspaceInvitation(dto *sendgrid.WorkspaceInvitationMailRequestDto) *restErrors.RestErr {
	return WorkspaceInvitationFunc(dto)
}

func newFiberCtx(dto interface{}, method func(c *fiber.Ctx) error, locals map[string]interface{}) ([]byte, *http.Response) {
	app := fiber.New()
	app.Post("/test/", func(c *fiber.Ctx) error {
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
	mailService = &mailServiceMock{}
	userService = &userServiceMock{}
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

		CountWorkspaceByUserIdFunc = func(userId string) (int64, *restErrors.RestErr) {
			return 2, nil
		}
		DeleteWorkspaceFunc = func(workspace *workspace.Workspace) *restErrors.RestErr {
			return nil
		}

		DeleteNamespaceFunc = func(name string) *restErrors.RestErr {
			return nil
		}

		_, resp := newFiberCtx("", Delete, locals)
		assert.EqualValues(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("Delete_Workspace_should_throw_if_user_has_only_one_workspace", func(t *testing.T) {
		CountWorkspaceByUserIdFunc = func(userId string) (int64, *restErrors.RestErr) {
			return 1, nil
		}

		body, resp := newFiberCtx("", Delete, locals)
		var restErr restErrors.RestErr
		err := json.Unmarshal(body, &restErr)
		assert.Nil(t, err)
		assert.EqualValues(t, "request declined, you should have at least 1 workspace!", restErr.Message)
		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Delete_Workspace_should_throw_if_count_user_workspace_throw", func(t *testing.T) {
		CountWorkspaceByUserIdFunc = func(userId string) (int64, *restErrors.RestErr) {
			return 0, restErrors.NewInternalServerError("something went wrong")
		}

		body, resp := newFiberCtx("", Delete, locals)
		var restErr restErrors.RestErr
		err := json.Unmarshal(body, &restErr)
		assert.Nil(t, err)
		assert.EqualValues(t, "something went wrong", restErr.Message)
		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("Delete_Workspace_should_throw_if_workspace_repo_throws", func(t *testing.T) {
		CountWorkspaceByUserIdFunc = func(userId string) (int64, *restErrors.RestErr) {
			return 2, nil
		}

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
		CountWorkspaceByUserIdFunc = func(userId string) (int64, *restErrors.RestErr) {
			return 2, nil
		}

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
		GetWorkspaceByUserIdFunc = func(userId string) ([]*workspace.Workspace, *restErrors.RestErr) {
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
		GetWorkspaceByUserIdFunc = func(userId string) ([]*workspace.Workspace, *restErrors.RestErr) {
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

func TestAddMemberToWorkspace(t *testing.T) {
	userDetails := new(token.UserDetails)
	userDetails.ID = "11"

	workspaceModelLocals := new(workspace.Workspace)
	workspaceUserModelLocals := new(workspaceuser.WorkspaceUser)
	workspaceUserModelLocals.UserId = userDetails.ID
	workspaceModelLocals.WorkspaceUsers = []workspaceuser.WorkspaceUser{*workspaceUserModelLocals}

	var locals = map[string]interface{}{}
	locals["user"] = *userDetails
	locals["workspace"] = *workspaceModelLocals

	var validDto = map[string]string{
		"email": "test@test.com",
		"role":  "admin",
	}
	var invalidDto = map[string]string{
		"email": "invalid",
		"role":  "invalid",
	}

	t.Run("add_member_to_workspace_should_pass", func(t *testing.T) {
		GetByEmailFunc = func(email string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}

		WorkspaceInvitationFunc = func(dto *sendgrid.WorkspaceInvitationMailRequestDto) *restErrors.RestErr {
			return nil
		}

		AddWorkspaceMemberFunc = func(workspace *workspace.Workspace, memberId string, role string) *restErrors.RestErr {
			return nil
		}

		result, resp := newFiberCtx(validDto, AddMember, locals)
		var responseMessage map[string]shared.SuccessMessage
		err := json.Unmarshal(result, &responseMessage)
		if err != nil {
			panic(err)
		}
		assert.EqualValues(t, "user has been added to the workspace", responseMessage["data"].Message)
		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("add_member_to_workspace_should_throw_validation_err", func(t *testing.T) {
		body, resp := newFiberCtx(invalidDto, AddMember, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}
		var fields = map[string]string{}
		fields["email"] = "email should be a valid email address"
		fields["role"] = "invalid role"
		badReqErr := restErrors.NewValidationError(fields)

		assert.EqualValues(t, badReqErr.Status, resp.StatusCode)
		assert.Equal(t, *badReqErr, result)
	})

	t.Run("Update_workspace_Should_Throw_Invalid_Request_Error", func(t *testing.T) {
		body, resp := newFiberCtx("", AddMember, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}
		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "invalid request body", result.Message)
	})

	t.Run("add_member_to_workspace_should_throw_if_user_already_a_member_of_the_workspace", func(t *testing.T) {
		GetByEmailFunc = func(email string) (*user.User, *restErrors.RestErr) {
			member := new(user.User)
			member.ID = userDetails.ID
			return member, nil
		}

		WorkspaceInvitationFunc = func(dto *sendgrid.WorkspaceInvitationMailRequestDto) *restErrors.RestErr {
			return nil
		}

		AddWorkspaceMemberFunc = func(workspace *workspace.Workspace, memberId string, role string) *restErrors.RestErr {
			return nil
		}

		result, _ := newFiberCtx(validDto, AddMember, locals)
		var restErr restErrors.RestErr
		err := json.Unmarshal(result, &restErr)
		if err != nil {
			panic(err)
		}
		assert.EqualValues(t, "User is already a member of the workspace", restErr.Message)
		assert.EqualValues(t, http.StatusConflict, restErr.Status)
	})

	t.Run("add_member_to_workspace_should_throw_if_member_does'not_exit", func(t *testing.T) {
		GetByEmailFunc = func(email string) (*user.User, *restErrors.RestErr) {
			return nil, restErrors.NewInternalServerError("something went wrong")
		}

		body, resp := newFiberCtx(validDto, AddMember, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err)
		}
		assert.EqualValues(t, http.StatusInternalServerError, result.Status)
		assert.EqualValues(t, "something went wrong", result.Message)
		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)
	})
	t.Run("add_member_to_workspace_should_throw_if_workspace_service_throw", func(t *testing.T) {
		GetByEmailFunc = func(email string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}

		WorkspaceInvitationFunc = func(dto *sendgrid.WorkspaceInvitationMailRequestDto) *restErrors.RestErr {
			return nil
		}

		AddWorkspaceMemberFunc = func(workspace *workspace.Workspace, memberId string, role string) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		body, resp := newFiberCtx(validDto, AddMember, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err)
		}
		assert.EqualValues(t, http.StatusInternalServerError, result.Status)
		assert.EqualValues(t, "something went wrong", result.Message)
		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)
	})

}

func TestLeaveWorkspace(t *testing.T) {
	userDetails := new(token.UserDetails)
	userDetails.ID = "11"

	workspaceModelLocals := new(workspace.Workspace)
	workspaceModelLocals.UserId = "ownerId"

	workspaceUserModelLocals := new(workspaceuser.WorkspaceUser)
	workspaceUserModelLocals.UserId = userDetails.ID
	workspaceUserModelLocals.Role = roles.Reader
	workspaceModelLocals.WorkspaceUsers = []workspaceuser.WorkspaceUser{*workspaceUserModelLocals}

	var locals = map[string]interface{}{}
	locals["user"] = *userDetails
	locals["workspace"] = *workspaceModelLocals
	locals["workspaceUser"] = *workspaceUserModelLocals

	t.Run("leave_workspace_should_pass", func(t *testing.T) {
		DeleteWorkspaceMemberFunc = func(workspace *workspace.Workspace, memberId string) *restErrors.RestErr {
			return nil
		}
		result, resp := newFiberCtx("", Leave, locals)
		var responseMessage map[string]shared.SuccessMessage
		err := json.Unmarshal(result, &responseMessage)
		if err != nil {
			panic(err)
		}
		assert.EqualValues(t, "You're no longer member of this workspace", responseMessage["data"].Message)
		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("leave_workspace_should_throw_if_the_admin_tries_to_leave_and_there_is_no_other_admin_in_the_work_space", func(t *testing.T) {
		workspaceUserModelLocalsIsAdmin := workspaceUserModelLocals
		workspaceUserModelLocalsIsAdmin.Role = roles.Admin
		locals["workspaceUser"] = *workspaceUserModelLocalsIsAdmin

		DeleteWorkspaceMemberFunc = func(workspace *workspace.Workspace, memberId string) *restErrors.RestErr {
			return nil
		}
		result, resp := newFiberCtx("", Leave, locals)
		var restErr restErrors.RestErr
		err := json.Unmarshal(result, &restErr)
		if err != nil {
			panic(err)
		}
		assert.EqualValues(t, "user with admin role can leave workspace only if it has another admin", restErr.Message)
		assert.EqualValues(t, http.StatusForbidden, resp.StatusCode)
		workspaceUserModelLocals.Role = roles.Reader
		locals["workspaceUser"] = *workspaceUserModelLocals

	})

	t.Run("leave_workspace_should_throw_if_service_Throw", func(t *testing.T) {
		DeleteWorkspaceMemberFunc = func(workspace *workspace.Workspace, memberId string) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}
		result, resp := newFiberCtx("", Leave, locals)
		var restErr restErrors.RestErr
		err := json.Unmarshal(result, &restErr)
		if err != nil {
			panic(err)
		}
		assert.EqualValues(t, "something went wrong", restErr.Message)
		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)
	})

}

func TestRemoveMemberWorkspace(t *testing.T) {
	userDetails := new(token.UserDetails)
	userDetails.ID = "11"

	workspaceModelLocals := new(workspace.Workspace)
	workspaceModelLocals.UserId = "11"

	workspaceUserModelLocals := new(workspaceuser.WorkspaceUser)
	workspaceUserModelLocals.UserId = ""
	workspaceModelLocals.WorkspaceUsers = []workspaceuser.WorkspaceUser{*workspaceUserModelLocals}

	var locals = map[string]interface{}{}
	locals["user"] = *userDetails
	locals["workspace"] = *workspaceModelLocals

	t.Run("remove_workspace_user_should_pass", func(t *testing.T) {
		DeleteWorkspaceMemberFunc = func(workspace *workspace.Workspace, memberId string) *restErrors.RestErr {
			return nil
		}

		result, resp := newFiberCtx("", RemoveMember, locals)
		var responseMessage map[string]shared.SuccessMessage
		err := json.Unmarshal(result, &responseMessage)
		if err != nil {
			panic(err)
		}

		assert.EqualValues(t, "User has been removed from workspace", responseMessage["data"].Message)
		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("user_can't_remove_him_self_from_workspace", func(t *testing.T) {
		userDetails.ID = ""
		locals["user"] = *userDetails
		workspaceUserModelLocals.UserId = userDetails.ID
		workspaceModelLocals.WorkspaceUsers = []workspaceuser.WorkspaceUser{*workspaceUserModelLocals}
		locals["workspace"] = *workspaceModelLocals

		DeleteWorkspaceMemberFunc = func(workspace *workspace.Workspace, memberId string) *restErrors.RestErr {
			return nil
		}

		result, resp := newFiberCtx("", RemoveMember, locals)
		var restErr restErrors.RestErr
		err := json.Unmarshal(result, &restErr)
		if err != nil {
			panic(err)
		}

		assert.EqualValues(t, "you can't remove your self, try to leave workspace instead!", restErr.Message)
		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		userDetails.ID = "11"
		locals["user"] = *userDetails
		workspaceUserModelLocals.UserId = ""
		workspaceModelLocals.WorkspaceUsers = []workspaceuser.WorkspaceUser{*workspaceUserModelLocals}
		locals["workspace"] = *workspaceModelLocals

	})

	t.Run("remove_member_should_throw_if_user_doesnt'_exits", func(t *testing.T) {
		userDetails.ID = "11"
		locals["user"] = *userDetails
		workspaceUserModelLocals.UserId = "12"
		workspaceModelLocals.WorkspaceUsers = []workspaceuser.WorkspaceUser{*workspaceUserModelLocals}
		locals["workspace"] = *workspaceModelLocals

		DeleteWorkspaceMemberFunc = func(workspace *workspace.Workspace, memberId string) *restErrors.RestErr {
			return nil
		}

		result, resp := newFiberCtx("", RemoveMember, locals)
		var restErr restErrors.RestErr
		err := json.Unmarshal(result, &restErr)
		if err != nil {
			panic(err)
		}

		assert.EqualValues(t, "user isn't a member of the workspace", restErr.Message)
		assert.EqualValues(t, http.StatusNotFound, resp.StatusCode)
		userDetails.ID = "11"
		locals["user"] = *userDetails
		workspaceUserModelLocals.UserId = ""
		workspaceModelLocals.WorkspaceUsers = []workspaceuser.WorkspaceUser{*workspaceUserModelLocals}
		locals["workspace"] = *workspaceModelLocals

	})

	t.Run("leave_workspace_should_throw_if_service_Throw", func(t *testing.T) {
		DeleteWorkspaceMemberFunc = func(workspace *workspace.Workspace, memberId string) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}
		result, resp := newFiberCtx("", RemoveMember, locals)
		var restErr restErrors.RestErr
		err := json.Unmarshal(result, &restErr)
		if err != nil {
			panic(err)
		}
		assert.EqualValues(t, "something went wrong", restErr.Message)
		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)
	})

}

func TestMembers(t *testing.T) {
	userDetails := new(token.UserDetails)
	userDetails.ID = uuid.NewString()

	workspaceModelLocals := new(workspace.Workspace)
	workspaceUserModelLocals := new(workspaceuser.WorkspaceUser)
	workspaceUserModelLocals.UserId = userDetails.ID
	workspaceModelLocals.WorkspaceUsers = []workspaceuser.WorkspaceUser{*workspaceUserModelLocals}

	var locals = map[string]interface{}{}
	locals["user"] = *userDetails
	locals["workspace"] = *workspaceModelLocals

	t.Run("list_workspace_members_should_pass", func(t *testing.T) {
		FindWhereIdInSliceFunc = func(ids []string) ([]*user.User, *restErrors.RestErr) {
			user1 := new(user.User)
			user1.Email = "email@test.com"
			user1.ID = uuid.NewString()
			return []*user.User{user1}, nil
		}

		body, resp := newFiberCtx("", Members, locals)
		var result map[string][]user.PublicUserResponseDto
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err)
		}

		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
		assert.Len(t, result["data"], 1)
		assert.EqualValues(t, result["data"][0].Email, "email@test.com")
	})

	t.Run("list_workspace_members_should_throw_if_service_throw", func(t *testing.T) {
		FindWhereIdInSliceFunc = func(ids []string) ([]*user.User, *restErrors.RestErr) {
			return nil, restErrors.NewInternalServerError("something went wrong")
		}

		body, resp := newFiberCtx("", Members, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err)
		}

		assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)
		assert.EqualValues(t, http.StatusInternalServerError, result.Status)
		assert.EqualValues(t, "something went wrong", result.Message)
	})
}
