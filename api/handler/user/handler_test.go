package user

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/kotalco/cloud-api/internal/workspace"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	"github.com/kotalco/cloud-api/pkg/token"
	"gorm.io/gorm"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kotalco/cloud-api/pkg/sendgrid"

	"github.com/gofiber/fiber/v2"

	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/cloud-api/internal/user"
	"github.com/kotalco/cloud-api/internal/verification"
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
Verification service Mocks
*/
var (
	CreateFunc                      func(userId string) (string, *restErrors.RestErr)
	GetByUserIdFunc                 func(userId string) (*verification.Verification, *restErrors.RestErr)
	ResendFunc                      func(userId string) (string, *restErrors.RestErr)
	VerifyFunc                      func(userId string, token string) *restErrors.RestErr
	VerificationWithTransactionFunc func(txHandle *gorm.DB) verification.IService
)

type verificationServiceMock struct{}

func (vService verificationServiceMock) WithTransaction(txHandle *gorm.DB) verification.IService {
	return vService
}

func (verificationServiceMock) Create(userId string) (string, *restErrors.RestErr) {
	return CreateFunc(userId)
}

func (verificationServiceMock) GetByUserId(userId string) (*verification.Verification, *restErrors.RestErr) {
	return GetByUserIdFunc(userId)
}

func (verificationServiceMock) Verify(userId string, token string) *restErrors.RestErr {
	return VerifyFunc(userId, token)
}

func (vService verificationServiceMock) Resend(userId string) (string, *restErrors.RestErr) {
	return ResendFunc(userId)
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

/*
Workspace service Mocks
*/
var (
	WorkspaceWithTransaction  func(txHandle *gorm.DB) workspace.IService
	CreateWorkspaceFunc       func(dto *workspace.CreateWorkspaceRequestDto, userId string) (*workspace.Workspace, *restErrors.RestErr)
	UpdateWorkspaceFunc       func(dto *workspace.UpdateWorkspaceRequestDto, workspace *workspace.Workspace) *restErrors.RestErr
	DeleteWorkspace           func(workspace *workspace.Workspace) *restErrors.RestErr
	GetWorkspaceByIdFunc      func(Id string) (*workspace.Workspace, *restErrors.RestErr)
	GetWorkspacesByUserIdFunc func(userId string) ([]*workspace.Workspace, *restErrors.RestErr)
	addWorkspaceMemberFunc    func(workspace *workspace.Workspace, memberId string) *restErrors.RestErr
	DeleteWorkspaceMemberFunc func(workspace *workspace.Workspace, memberId string) *restErrors.RestErr
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

func (workspaceServiceMock) Delete(workspace *workspace.Workspace) *restErrors.RestErr {
	return DeleteWorkspace(workspace)
}

func (workspaceServiceMock) GetById(id string) (*workspace.Workspace, *restErrors.RestErr) {
	return GetWorkspaceByIdFunc(id)
}

func (workspaceServiceMock) GetByUserId(userId string) ([]*workspace.Workspace, *restErrors.RestErr) {
	return GetWorkspacesByUserIdFunc(userId)
}

func (workspaceServiceMock) AddWorkspaceMember(workspace *workspace.Workspace, memberId string) *restErrors.RestErr {
	return addWorkspaceMemberFunc(workspace, memberId)
}

func (workspaceServiceMock) DeleteWorkspaceMember(workspace *workspace.Workspace, memberId string) *restErrors.RestErr {
	return DeleteWorkspaceMemberFunc(workspace, memberId)
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

func TestMain(m *testing.M) {
	userService = &userServiceMock{}
	verificationService = &verificationServiceMock{}
	mailService = &mailServiceMock{}
	workspaceService = &workspaceServiceMock{}
	namespaceService = &namespaceServiceMock{}

	sqlclient.OpenDBConnection()

	code := m.Run()
	os.Exit(code)
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

func TestSignUp(t *testing.T) {
	var validDto = map[string]string{
		"email":                 "test@test.com",
		"password":              "123456",
		"password_confirmation": "123456",
	}
	var invalidDto = map[string]string{
		"email":                 "testtest.com",
		"password":              "1",
		"password_confirmation": "123",
	}

	t.Run("Sign_Up_Should_Pass", func(t *testing.T) {
		SignUpFunc = func(dto *user.SignUpRequestDto) (*user.User, *restErrors.RestErr) {
			newUser := new(user.User)
			newUser.Email = "test@test.com"
			return newUser, nil
		}

		CreateFunc = func(userId string) (string, *restErrors.RestErr) {
			return "JWT-token", nil
		}

		CreateWorkspaceFunc = func(dto *workspace.CreateWorkspaceRequestDto, userId string) (*workspace.Workspace, *restErrors.RestErr) {
			responseDto := new(workspace.Workspace)
			responseDto.ID = uuid.New().String()
			responseDto.Name = "testNamespace"
			responseDto.K8sNamespace = "testNamespace" + "-" + responseDto.ID
			return responseDto, nil
		}

		CreateNamespaceFunc = func(name string) *restErrors.RestErr {
			return nil
		}

		SignUpMailFunc = func(dto *sendgrid.MailRequestDto) *restErrors.RestErr {
			return nil
		}

		body, resp := newFiberCtx(validDto, SignUp, map[string]interface{}{})

		var result map[string]user.UserResponseDto
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusCreated, resp.StatusCode)
		assert.EqualValues(t, "test@test.com", result["data"].Email)
	})

	t.Run("Sign_Up_Should_Pass_Even_If_Can't_Create_Work_Space", func(t *testing.T) {
		SignUpFunc = func(dto *user.SignUpRequestDto) (*user.User, *restErrors.RestErr) {
			newUser := new(user.User)
			newUser.Email = "test@test.com"
			return newUser, nil
		}

		CreateFunc = func(userId string) (string, *restErrors.RestErr) {
			return "JWT-token", nil
		}

		CreateWorkspaceFunc = func(dto *workspace.CreateWorkspaceRequestDto, userId string) (*workspace.Workspace, *restErrors.RestErr) {
			return nil, restErrors.NewInternalServerError("can't create workspace")
		}

		SignUpMailFunc = func(dto *sendgrid.MailRequestDto) *restErrors.RestErr {
			return nil
		}

		body, resp := newFiberCtx(validDto, SignUp, map[string]interface{}{})

		var result map[string]user.UserResponseDto
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusCreated, resp.StatusCode)
		assert.EqualValues(t, "test@test.com", result["data"].Email)
	})
	t.Run("Sign_Up_Should_Pass_Even_If_Can't_Create_Namespace", func(t *testing.T) {
		SignUpFunc = func(dto *user.SignUpRequestDto) (*user.User, *restErrors.RestErr) {
			newUser := new(user.User)
			newUser.Email = "test@test.com"
			return newUser, nil
		}

		CreateFunc = func(userId string) (string, *restErrors.RestErr) {
			return "JWT-token", nil
		}

		CreateWorkspaceFunc = func(dto *workspace.CreateWorkspaceRequestDto, userId string) (*workspace.Workspace, *restErrors.RestErr) {
			responseDto := new(workspace.Workspace)
			responseDto.ID = uuid.New().String()
			responseDto.Name = "testNamespace"
			responseDto.K8sNamespace = "testNamespace" + "-" + responseDto.ID
			return responseDto, nil
		}

		CreateNamespaceFunc = func(name string) *restErrors.RestErr {
			return restErrors.NewInternalServerError("can't create namespace")
		}

		SignUpMailFunc = func(dto *sendgrid.MailRequestDto) *restErrors.RestErr {
			return nil
		}

		body, resp := newFiberCtx(validDto, SignUp, map[string]interface{}{})

		var result map[string]user.UserResponseDto
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusCreated, resp.StatusCode)
		assert.EqualValues(t, "test@test.com", result["data"].Email)
	})

	t.Run("Sign_Up_Should_Throw_Validation_Errors", func(t *testing.T) {

		body, resp := newFiberCtx(invalidDto, SignUp, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		fields := map[string]string{}
		fields["email"] = "invalid email address"
		fields["password"] = "password should be at least 6 chars"
		fields["password_confirmation"] = "password_confirmation should match password field"
		badReqErr := restErrors.NewValidationError(fields)

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.Equal(t, *badReqErr, result)
	})

	t.Run("Sing_Up_Should_Throw_Invalid_Request_Error", func(t *testing.T) {

		body, resp := newFiberCtx("", SignUp, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "invalid request body", result.Message)
	})

	t.Run("Sing_Up_Should_Throw_if_Service_Throws", func(t *testing.T) {
		SignUpFunc = func(dto *user.SignUpRequestDto) (*user.User, *restErrors.RestErr) {
			return nil, restErrors.NewBadRequestError("user service errors")
		}

		body, resp := newFiberCtx(validDto, SignUp, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, result.Message, "user service errors")
	})

	t.Run("Sign_Up_Should_Throw_if_Verification_Service_Throws", func(t *testing.T) {
		SignUpFunc = func(dto *user.SignUpRequestDto) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}

		CreateFunc = func(userId string) (string, *restErrors.RestErr) {
			return "", restErrors.NewBadRequestError("verification service errors")
		}

		body, resp := newFiberCtx(validDto, SignUp, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, result.Message, "verification service errors")
	})

}

func TestVerifyEmail(t *testing.T) {
	var validDto = map[string]string{
		"email": "test@test.com",
		"token": "UvMIGpBdgfYoRrhkJmTiKWzjUrXdLXihWNVssiNUXLuiokXlwRRsfFcyqaWzsCSmurPNKOBFhPmBRRZd",
	}
	var invalidDto = map[string]string{
		"email": "testcom",
		"token": "Uv",
	}

	t.Run("Verify_Email_Should_Pass", func(t *testing.T) {
		GetByEmailFunc = func(email string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}

		VerifyEmailFunc = func(model *user.User) *restErrors.RestErr {
			return nil
		}

		VerifyFunc = func(userId string, token string) *restErrors.RestErr {
			return nil
		}

		body, resp := newFiberCtx(validDto, VerifyEmail, map[string]interface{}{})

		type message struct {
			Message string
		}

		var result map[string]message
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
		assert.EqualValues(t, result["data"].Message, "email verified")
	})

	t.Run("Verify_Email_Should_Throw_Validation_Errors", func(t *testing.T) {
		body, resp := newFiberCtx(invalidDto, VerifyEmail, map[string]interface{}{})
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		fields := map[string]string{}
		fields["email"] = "invalid email address"
		fields["token"] = "invalid token signature"
		badReqErr := restErrors.NewValidationError(fields)

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.Equal(t, *badReqErr, result)
	})

	t.Run("Verify_Email_Should_Throw_Invalid_Request_Error", func(t *testing.T) {
		body, resp := newFiberCtx("", VerifyEmail, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "invalid request body", result.Message)
	})

	t.Run("Verify_Email_Should_Throw_No_Such_Email", func(t *testing.T) {
		GetByEmailFunc = func(email string) (*user.User, *restErrors.RestErr) {
			return nil, restErrors.NewNotFoundError(fmt.Sprintf("can't find user with email  %s", email))
		}

		body, resp := newFiberCtx(validDto, VerifyEmail, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusNotFound, resp.StatusCode)
		assert.EqualValues(t, result.Message, "can't find user with email  test@test.com")
	})

	t.Run("Verify_Email_Should_Throw_Email_Already_Verified", func(t *testing.T) {
		GetByEmailFunc = func(email string) (*user.User, *restErrors.RestErr) {
			newUser := new(user.User)
			newUser.IsEmailVerified = true
			return newUser, nil
		}

		body, resp := newFiberCtx(validDto, VerifyEmail, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, result.Message, "email already verified")
	})

	t.Run("Verify_Email_Should_Throw_If_User_Service_Throws", func(t *testing.T) {
		GetByEmailFunc = func(email string) (*user.User, *restErrors.RestErr) {
			newUser := new(user.User)
			newUser.Email = "test@test.com"
			return newUser, nil
		}

		VerifyFunc = func(userId string, token string) *restErrors.RestErr {
			return nil
		}

		VerifyEmailFunc = func(model *user.User) *restErrors.RestErr {
			return restErrors.NewBadRequestError("user service error")
		}

		body, resp := newFiberCtx(validDto, VerifyEmail, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, result.Message, "user service error")
	})

	t.Run("Verify_Email_Should_Throw_If_Verification_Service_Throws", func(t *testing.T) {
		GetByEmailFunc = func(email string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}

		VerifyFunc = func(userId string, token string) *restErrors.RestErr {
			return restErrors.NewBadRequestError("verification service error")
		}

		body, resp := newFiberCtx(validDto, VerifyEmail, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, result.Message, "verification service error")
	})
}

func TestSignIn(t *testing.T) {
	validDto := map[string]string{
		"email":    "test@test.com",
		"password": "123456",
	}
	invalidDto := map[string]string{
		"email":    "testtestcom",
		"password": "1",
	}

	t.Run("Sing_In_Should_Pass", func(t *testing.T) {
		SignInFunc = func(dto *user.SignInRequestDto) (*user.UserSessionResponseDto, *restErrors.RestErr) {
			session := new(user.UserSessionResponseDto)
			session.Token = "token"
			session.Authorized = true
			return session, nil
		}

		body, resp := newFiberCtx(validDto, SignIn, map[string]interface{}{})
		var result map[string]user.UserSessionResponseDto
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
		assert.EqualValues(t, "token", result["data"].Token)
		assert.EqualValues(t, true, result["data"].Authorized)
	})

	t.Run("Sing_In_Should_throw_Validation_Error", func(t *testing.T) {
		SignInFunc = func(dto *user.SignInRequestDto) (*user.UserSessionResponseDto, *restErrors.RestErr) {
			return new(user.UserSessionResponseDto), nil
		}

		body, resp := newFiberCtx(invalidDto, SignIn, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		fields := map[string]string{}
		fields["email"] = "invalid email address"
		fields["password"] = "password should be at least 6 chars"
		badReqErr := restErrors.NewValidationError(fields)

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.Equal(t, *badReqErr, result)
	})

	t.Run("Sign_In_Should_throw_Invalid_Request_Body", func(t *testing.T) {
		body, resp := newFiberCtx("", SignIn, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		badReqErr := restErrors.NewBadRequestError("invalid request body")

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, badReqErr.Message, result.Message)
	})

	t.Run("Sing_In_Should_throw_if_user_service_throws", func(t *testing.T) {
		SignInFunc = func(dto *user.SignInRequestDto) (*user.UserSessionResponseDto, *restErrors.RestErr) {
			return nil, restErrors.NewBadRequestError("error from service")
		}

		body, resp := newFiberCtx(validDto, SignIn, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.Equal(t, "error from service", result.Message)
	})
}

func TestSendEmailVerification(t *testing.T) {
	validDto := map[string]string{"email": "test@test.com"}
	invalidDto := map[string]string{"email": "test"}

	t.Run("Send_Email_Verification_Should_Pass", func(t *testing.T) {
		GetByEmailFunc = func(email string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}

		ResendFunc = func(userId string) (string, *restErrors.RestErr) {
			return "token", nil
		}

		ResendEmailVerificationFunc = func(dto *sendgrid.MailRequestDto) *restErrors.RestErr {
			return nil
		}

		body, resp := newFiberCtx(validDto, SendEmailVerification, map[string]interface{}{})

		type message struct {
			Message string
		}

		var result map[string]message
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
		assert.EqualValues(t, "email verification sent successfully", result["data"].Message)
	})

	t.Run("Send_Email_Verification_Should_Throw_Validation_Errors", func(t *testing.T) {
		GetByEmailFunc = func(email string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}

		body, resp := newFiberCtx(invalidDto, SendEmailVerification, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		fields := map[string]string{}
		fields["email"] = "invalid email address"
		badReqErr := restErrors.NewValidationError(fields)

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.Equal(t, *badReqErr, result)
	})

	t.Run("Send_Email_Verification_Should_Throw_Invalid_Request_Error", func(t *testing.T) {
		body, resp := newFiberCtx("", SendEmailVerification, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		badReqErr := restErrors.NewBadRequestError("invalid request body")

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, badReqErr.Message, result.Message)
	})

	t.Run("Send_Email_Verification_Should_Throw_Email_Already_Verified", func(t *testing.T) {
		GetByEmailFunc = func(email string) (*user.User, *restErrors.RestErr) {
			newUser := new(user.User)
			newUser.IsEmailVerified = true
			return newUser, nil
		}

		body, resp := newFiberCtx(validDto, SendEmailVerification, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "email already verified", result.Message)
	})

	t.Run("Send_Email_Verification_Should_Throw_If_User_Service_Throws", func(t *testing.T) {
		GetByEmailFunc = func(email string) (*user.User, *restErrors.RestErr) {
			return nil, restErrors.NewBadRequestError("user service error")
		}

		body, resp := newFiberCtx(validDto, SendEmailVerification, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "user service error", result.Message)
	})

	t.Run("Send_Email_Verification_Should_Throw_If_Verification_Service_Throws", func(t *testing.T) {
		GetByEmailFunc = func(email string) (*user.User, *restErrors.RestErr) {
			newUser := new(user.User)
			newUser.Email = "test@test.com"
			return newUser, nil
		}

		ResendFunc = func(userId string) (string, *restErrors.RestErr) {
			return "", restErrors.NewBadRequestError("verification service errors")
		}

		body, resp := newFiberCtx(validDto, SendEmailVerification, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "verification service errors", result.Message)
	})
}

func TestForgetPassword(t *testing.T) {
	t.Parallel()
	validDto := map[string]string{"email": "test@test.com"}
	invalidDto := map[string]string{"email": "test"}

	t.Run("Forget_Password_Should_Pass", func(t *testing.T) {
		GetByEmailFunc = func(email string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}
		ResendFunc = func(userId string) (string, *restErrors.RestErr) {
			return "token", nil
		}
		ForgetPasswordMailFunc = func(dto *sendgrid.MailRequestDto) *restErrors.RestErr {
			return nil
		}
		body, resp := newFiberCtx(validDto, ForgetPassword, map[string]interface{}{})

		type message struct {
			Message string
		}

		var result map[string]message
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
		assert.EqualValues(t, "reset password has been sent to your email", result["data"].Message)
	})

	t.Run("Forget_Password_Should_Throw_validation_Errors", func(t *testing.T) {
		body, resp := newFiberCtx(invalidDto, ForgetPassword, map[string]interface{}{})

		var result restErrors.RestErr

		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		fields := map[string]string{
			"email": "invalid email address",
		}
		badReqErr := restErrors.NewValidationError(fields)

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, *badReqErr, result)
	})

	t.Run("Forget_Password_Should_Throw_Invalid_Request_Body", func(t *testing.T) {
		body, resp := newFiberCtx("", ForgetPassword, map[string]interface{}{})

		var result restErrors.RestErr

		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "invalid request body", result.Message)
	})

	t.Run("Forget_Password_Should_Throw_If_User_Service_Throws", func(t *testing.T) {
		GetByEmailFunc = func(email string) (*user.User, *restErrors.RestErr) {
			return nil, restErrors.NewBadRequestError("user service errors")
		}

		body, resp := newFiberCtx(validDto, ForgetPassword, map[string]interface{}{})

		var result restErrors.RestErr

		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}
		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "user service errors", result.Message)
	})

	t.Run("Forget_Password_Should_Throw_If_Verification_Service_Throws", func(t *testing.T) {
		GetByEmailFunc = func(email string) (*user.User, *restErrors.RestErr) {
			newUser := new(user.User)
			newUser.Email = "test@test.com"
			return newUser, nil
		}

		ResendFunc = func(userId string) (string, *restErrors.RestErr) {
			return "", restErrors.NewBadRequestError("verification service errors")
		}

		body, resp := newFiberCtx(validDto, ForgetPassword, map[string]interface{}{})

		var result restErrors.RestErr

		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "verification service errors", result.Message)
	})
}

func TestResetPassword(t *testing.T) {
	validDto := map[string]string{
		"email":                 "test@test.com",
		"password":              "123456",
		"password_confirmation": "123456",
		"token":                 "UvMIGpBdgfYoRrhkJmTiKWzjUrXdLXihWNVssiNUXLuiokXlwRRsfFcyqaWzsCSmurPNKOBFhPmBRRZd",
	}
	invalidDto := map[string]string{
		"email":                 "testtestcom",
		"password":              "1",
		"password_confirmation": "12",
		"token":                 "123",
	}

	t.Run("Reset_Password_Should_Pass", func(t *testing.T) {
		GetByEmailFunc = func(email string) (*user.User, *restErrors.RestErr) {
			newUser := new(user.User)
			newUser.IsEmailVerified = true
			return newUser, nil
		}
		VerifyFunc = func(userId string, token string) *restErrors.RestErr {
			return nil
		}
		ResetPasswordFunc = func(model *user.User, password string) *restErrors.RestErr {
			return nil
		}

		body, resp := newFiberCtx(validDto, ResetPassword, map[string]interface{}{})

		type message struct {
			Message string
		}
		var result map[string]message

		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
		assert.EqualValues(t, "password reset successfully", result["data"].Message)
	})

	t.Run("Reset_Password_Should_Throw_Validation_Errors", func(t *testing.T) {
		body, resp := newFiberCtx(invalidDto, ResetPassword, map[string]interface{}{})

		var result restErrors.RestErr

		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		var fields = map[string]string{}
		fields["email"] = "invalid email address"
		fields["password"] = "password should be at least 6 chars"
		fields["password_confirmation"] = "password_confirmation should match password field"
		fields["token"] = "invalid token signature"

		badReqErr := restErrors.NewValidationError(fields)

		assert.EqualValues(t, badReqErr.Status, resp.StatusCode)
		assert.Equal(t, *badReqErr, result)
	})

	t.Run("Reset_Password_Should_Throw_Invalid_Request_Error", func(t *testing.T) {
		body, resp := newFiberCtx("", ResetPassword, map[string]interface{}{})

		var result restErrors.RestErr

		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "invalid request body", result.Message)
	})

	t.Run("Reset_Password_Should_Throw_Email_Not_Verified", func(t *testing.T) {
		GetByEmailFunc = func(email string) (*user.User, *restErrors.RestErr) {
			newUser := new(user.User)
			newUser.IsEmailVerified = false
			return newUser, nil
		}

		body, resp := newFiberCtx(validDto, ResetPassword, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusForbidden, resp.StatusCode)
		assert.EqualValues(t, "email not verified", result.Message)
	})

	t.Run("Reset_Password_Should_Throw_If_User_Service_throws", func(t *testing.T) {
		t.Run("Reset_Password_No_Such_Email", func(t *testing.T) {
			GetByEmailFunc = func(email string) (*user.User, *restErrors.RestErr) {
				return nil, restErrors.NewBadRequestError("no such email error")
			}

			body, resp := newFiberCtx(validDto, ResetPassword, map[string]interface{}{})

			var result restErrors.RestErr
			err := json.Unmarshal(body, &result)
			if err != nil {
				panic(err.Error())
			}

			assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
			assert.EqualValues(t, "no such email error", result.Message)
		})
		t.Run("Reset_Password_No_Cannot_Reset_Email", func(t *testing.T) {
			GetByEmailFunc = func(email string) (*user.User, *restErrors.RestErr) {
				newUser := new(user.User)
				newUser.IsEmailVerified = true
				return newUser, nil
			}

			VerifyFunc = func(userId string, token string) *restErrors.RestErr {
				return nil
			}

			ResetPasswordFunc = func(model *user.User, password string) *restErrors.RestErr {
				return restErrors.NewInternalServerError("can't reset password error")
			}

			body, resp := newFiberCtx(validDto, ResetPassword, map[string]interface{}{})

			var result restErrors.RestErr
			err := json.Unmarshal(body, &result)
			if err != nil {
				panic(err.Error())
			}

			assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode)
			assert.EqualValues(t, "can't reset password error", result.Message)
		})

	})

	t.Run("Reset_Password_Should_Throw_If_Verification_Service_Throws", func(t *testing.T) {
		GetByEmailFunc = func(email string) (*user.User, *restErrors.RestErr) {
			newUser := new(user.User)
			newUser.IsEmailVerified = true
			return newUser, nil
		}

		VerifyFunc = func(userId string, token string) *restErrors.RestErr {
			return restErrors.NewBadRequestError("verification service errors")
		}

		body, resp := newFiberCtx(validDto, ResetPassword, map[string]interface{}{})

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "verification service errors", result.Message)
	})
}

func TestChangePassword(t *testing.T) {
	validDto := map[string]string{
		"old_password":          "123456",
		"password":              "123456",
		"password_confirmation": "123456",
	}
	invalidDto := map[string]string{
		"old_password":          "1",
		"password":              "123",
		"password_confirmation": "1234",
	}

	userDetails := new(token.UserDetails)
	userDetails.ID = "test@test.com"
	var locals = map[string]interface{}{}
	locals["user"] = *userDetails

	t.Run("Change_Password_Should_Pass", func(t *testing.T) {
		ChangePasswordFunc = func(model *user.User, dto *user.ChangePasswordRequestDto) *restErrors.RestErr {
			return nil
		}
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}
		body, resp := newFiberCtx(validDto, ChangePassword, locals)

		type message struct {
			Message string
		}

		var result map[string]message
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
		assert.EqualValues(t, "password changed successfully", result["data"].Message)
	})

	t.Run("Change_Password_Should_Throw_Validation_Errors", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}
		body, resp := newFiberCtx(invalidDto, ChangePassword, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		var fields = map[string]string{}
		fields["old_password"] = "password should be at least 6 chars"
		fields["password"] = "password should be at least 6 chars"
		fields["password_confirmation"] = "password_confirmation should match password field"

		badReqErr := restErrors.NewValidationError(fields)

		assert.EqualValues(t, badReqErr.Status, resp.StatusCode)
		assert.Equal(t, *badReqErr, result)
	})

	t.Run("Change_Password_Should_Throw_Invalid_Request_Error", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}
		body, resp := newFiberCtx("", ChangePassword, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "invalid request body", result.Message)
	})

	t.Run("Change_Password_Should_Throw_If_User_Service_Throw", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}
		ChangePasswordFunc = func(model *user.User, dto *user.ChangePasswordRequestDto) *restErrors.RestErr {
			return restErrors.NewBadRequestError("user service error")
		}

		body, resp := newFiberCtx(validDto, ChangePassword, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "user service error", result.Message)
	})
	t.Run("Change_Password_Should_Throw_If_User_Does_not_exist", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return nil, restErrors.NewNotFoundError("no such user")
		}

		body, resp := newFiberCtx(validDto, ChangePassword, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusNotFound, resp.StatusCode)
		assert.EqualValues(t, "no such user", result.Message)
	})

}

func TestChangeEmail(t *testing.T) {
	validDto := map[string]string{
		"email":    "test@test.com",
		"password": "123456",
	}
	inValidDto := map[string]string{
		"email":    "testcom",
		"password": "123",
	}

	userDetails := new(token.UserDetails)
	userDetails.ID = "test@test.com"
	var locals = map[string]interface{}{}
	locals["user"] = *userDetails

	t.Run("Change_Email_Should_Pass", func(t *testing.T) {
		ChangeEmailFunc = func(model *user.User, dto *user.ChangeEmailRequestDto) *restErrors.RestErr {
			return nil
		}
		ResendFunc = func(userId string) (string, *restErrors.RestErr) {
			return "token", nil
		}
		ResendEmailVerificationFunc = func(dto *sendgrid.MailRequestDto) *restErrors.RestErr {
			return nil
		}
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}

		body, resp := newFiberCtx(validDto, ChangeEmail, locals)

		type message struct {
			Message string
		}
		var result map[string]message
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
		assert.EqualValues(t, "email changed successfully", result["data"].Message)

	})

	t.Run("Change_Email_Should_Throw_Validation_Errors", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}
		body, resp := newFiberCtx(inValidDto, ChangeEmail, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		var fields = map[string]string{}
		fields["email"] = "invalid email address"
		fields["password"] = "password should be at least 6 chars"

		badReqErr := restErrors.NewValidationError(fields)

		assert.EqualValues(t, badReqErr.Status, resp.StatusCode)
		assert.Equal(t, *badReqErr, result)
	})

	t.Run("Change_Email_Should_Throw_Invalid_Request_Error", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}
		body, resp := newFiberCtx("", ChangeEmail, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "invalid request body", result.Message)
	})

	t.Run("Change_Email_Should_Throw_If_User_Service_Throw", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}
		ChangeEmailFunc = func(model *user.User, dto *user.ChangeEmailRequestDto) *restErrors.RestErr {
			return restErrors.NewBadRequestError("change email user service error")
		}

		body, resp := newFiberCtx(validDto, ChangeEmail, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "change email user service error", result.Message)
	})
	t.Run("Change_Email_Should_Throw_If_Verification_Service_Throw", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}
		ChangeEmailFunc = func(model *user.User, dto *user.ChangeEmailRequestDto) *restErrors.RestErr {
			return nil
		}
		ResendFunc = func(userId string) (string, *restErrors.RestErr) {
			return "", restErrors.NewBadRequestError("verification service error")
		}

		body, resp := newFiberCtx(validDto, ChangeEmail, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "verification service error", result.Message)
	})
	t.Run("Change_Email_Should_Throw_If_User_Not_Found", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return nil, restErrors.NewNotFoundError("no such user")
		}

		body, resp := newFiberCtx(validDto, ChangeEmail, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusNotFound, resp.StatusCode)
		assert.EqualValues(t, "no such user", result.Message)
	})

}

func TestWhoami(t *testing.T) {
	t.Run("Whoami_Should_Pass", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			model := new(user.User)
			model.ID = "1"
			return model, nil
		}
		userDetails := new(token.UserDetails)
		userDetails.ID = "1"
		var locals = map[string]interface{}{}
		locals["user"] = *userDetails

		body, resp := newFiberCtx(new(interface{}), Whoami, locals)

		var result map[string]user.UserResponseDto

		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, userDetails.ID, result["data"].ID)
		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
	})
}

func TestCreateTOTP(t *testing.T) {
	userDetails := new(token.UserDetails)
	userDetails.ID = "1"
	var locals = map[string]interface{}{}
	locals["user"] = *userDetails

	dto := new(user.CreateTOTPRequestDto)
	dto.Password = "123456"

	t.Run("Create_TOTP_Should_Pass", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}
		CreateTOTPFunc = func(model *user.User, dto *user.CreateTOTPRequestDto) (bytes.Buffer, *restErrors.RestErr) {
			return bytes.Buffer{}, nil
		}

		_, resp := newFiberCtx(dto, CreateTOTP, locals)
		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
	})
	t.Run("Create_TOTP_Should_Throw_Invalid_Request_Body", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}
		_, resp := newFiberCtx("", CreateTOTP, locals)
		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Create_TOTP_Should_Throw_Validation_Errors", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}
		invalidDto := new(user.CreateTOTPRequestDto)
		invalidDto.Password = "123"
		body, resp := newFiberCtx(invalidDto, CreateTOTP, locals)

		var result restErrors.RestErr

		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}
		var fields = map[string]string{}
		fields["password"] = "password should be at least 6 chars"

		badReqErr := restErrors.NewValidationError(fields)

		assert.EqualValues(t, badReqErr.Status, resp.StatusCode)
		assert.Equal(t, *badReqErr, result)
	})

	t.Run("Create_TOTP_Should_Throw_If_User_Service_Throws", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}
		CreateTOTPFunc = func(model *user.User, dto *user.CreateTOTPRequestDto) (bytes.Buffer, *restErrors.RestErr) {
			return bytes.Buffer{}, restErrors.NewBadRequestError("user service errors")
		}

		body, resp := newFiberCtx(dto, CreateTOTP, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, result.Message, result.Message)
	})
	t.Run("Create_TOTP_Should_Throw_If_User_Not_Found", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return nil, restErrors.NewNotFoundError("no such user")
		}

		body, resp := newFiberCtx(dto, CreateTOTP, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusNotFound, resp.StatusCode)
		assert.EqualValues(t, "no such user", result.Message)
	})

}

func TestEnableTwoFactorAuth(t *testing.T) {
	userDetails := new(token.UserDetails)
	userDetails.ID = "1"
	var locals = map[string]interface{}{}
	locals["user"] = *userDetails

	validDto := map[string]string{
		"totop": "123456",
	}

	t.Run("Enable_Two_Factor_Auth_Should_Pass", func(t *testing.T) {

		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}
		EnableTwoFactorAuthFunc = func(model *user.User, totp string) (*user.User, *restErrors.RestErr) {
			user := new(user.User)
			user.TwoFactorEnabled = true
			return user, nil

		}

		body, resp := newFiberCtx(validDto, EnableTwoFactorAuth, locals)

		var result map[string]user.UserResponseDto
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Enable_Two_Factor_Auth_Should_Throw_Invalid_Request_Body", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}
		body, resp := newFiberCtx("", EnableTwoFactorAuth, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "invalid request body", result.Message)
	})

	t.Run("Enable_Two_Factor_Auth_Should_Throw_if_user_Service_Throws", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}
		EnableTwoFactorAuthFunc = func(model *user.User, totp string) (*user.User, *restErrors.RestErr) {
			return nil, restErrors.NewBadRequestError("user service errors")
		}

		body, resp := newFiberCtx(validDto, EnableTwoFactorAuth, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "user service errors", result.Message)
	})
	t.Run("Enable_Two_Factor_Auth_Should_Throw_if_user_Not_Found", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return nil, restErrors.NewNotFoundError("no such user")
		}

		body, resp := newFiberCtx(validDto, EnableTwoFactorAuth, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusNotFound, resp.StatusCode)
		assert.EqualValues(t, "no such user", result.Message)
	})

}

func TestVerifyTOTP(t *testing.T) {
	userDetails := new(token.UserDetails)
	userDetails.ID = "1"
	var locals = map[string]interface{}{}
	locals["user"] = *userDetails

	validDto := map[string]string{
		"totop": "123456",
	}

	t.Run("Verify_TOTP_Should_Pass", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}
		VerifyTOTPFunc = func(model *user.User, totp string) (*user.UserSessionResponseDto, *restErrors.RestErr) {
			return new(user.UserSessionResponseDto), nil
		}

		body, resp := newFiberCtx(validDto, VerifyTOTP, locals)

		var result map[string]user.UserSessionResponseDto
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Verify_TOTP_Should_Throw_Invalid_Request_Body", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}
		body, resp := newFiberCtx("", VerifyTOTP, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "invalid request body", result.Message)
	})

	t.Run("Verify_TOTP_Should_Throw_If_User_Service_Throws", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}
		VerifyTOTPFunc = func(model *user.User, totp string) (*user.UserSessionResponseDto, *restErrors.RestErr) {
			return nil, restErrors.NewBadRequestError("user service can't verify otp")
		}

		body, resp := newFiberCtx(validDto, VerifyTOTP, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "user service can't verify otp", result.Message)
	})
	t.Run("Verify_TOTP_Should_Throw_If_User_Not_Found", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return nil, restErrors.NewNotFoundError("no such user")
		}

		body, resp := newFiberCtx(validDto, VerifyTOTP, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusNotFound, resp.StatusCode)
		assert.EqualValues(t, "no such user", result.Message)
	})

}

func TestDisableTwoFactorAuth(t *testing.T) {
	userDetails := new(token.UserDetails)
	userDetails.ID = "1"
	var locals = map[string]interface{}{}
	locals["user"] = *userDetails

	dto := new(user.DisableTOTPRequestDto)
	dto.Password = "123456"

	t.Run("Disable_Two_Factor_Auth_Should_Pass", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}
		DisableTwoFactorAuthFunc = func(model *user.User, dto *user.DisableTOTPRequestDto) *restErrors.RestErr {
			return nil
		}

		body, resp := newFiberCtx(dto, DisableTwoFactorAuth, locals)

		type message struct {
			Message string
		}
		var result map[string]message
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
		assert.EqualValues(t, "2FA disabled", result["data"].Message)
	})
	t.Run("Disable_Two_Factor_Auth_Should_Throw_Invalid_Request_Body", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}

		body, resp := newFiberCtx("", DisableTwoFactorAuth, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, http.StatusBadRequest, result.Status)
		assert.EqualValues(t, "invalid request body", result.Message)
	})

	t.Run("Disable_Two_Factor_Auth_Should_Throw_validation_Errors", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}

		DisableTwoFactorAuthFunc = func(model *user.User, dto *user.DisableTOTPRequestDto) *restErrors.RestErr {
			return nil
		}
		invalidDto := new(user.DisableTOTPRequestDto)
		invalidDto.Password = "123"

		body, resp := newFiberCtx(invalidDto, DisableTwoFactorAuth, locals)
		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		var fields = map[string]string{}
		fields["password"] = "password should be at least 6 chars"

		badReqErr := restErrors.NewValidationError(fields)

		assert.EqualValues(t, badReqErr.Status, resp.StatusCode)
		assert.Equal(t, *badReqErr, result)
	})

	t.Run("Disable_Two_Factor_Auth_Should_Throw_If_2Fa_Already_Disable", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return new(user.User), nil
		}

		DisableTwoFactorAuthFunc = func(model *user.User, dto *user.DisableTOTPRequestDto) *restErrors.RestErr {
			return restErrors.NewBadRequestError("2fa already disabled")
		}

		body, resp := newFiberCtx(dto, DisableTwoFactorAuth, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, "2fa already disabled", result.Message)
	})

	t.Run("Disable_Two_Factor_Auth_Should_Throw_If_user_not_found", func(t *testing.T) {
		GetByIdFunc = func(Id string) (*user.User, *restErrors.RestErr) {
			return nil, restErrors.NewNotFoundError("no such user")
		}

		DisableTwoFactorAuthFunc = func(model *user.User, dto *user.DisableTOTPRequestDto) *restErrors.RestErr {
			return restErrors.NewBadRequestError("2fa already disabled")
		}

		body, resp := newFiberCtx(dto, DisableTwoFactorAuth, locals)

		var result restErrors.RestErr
		err := json.Unmarshal(body, &result)
		if err != nil {
			panic(err.Error())
		}

		assert.EqualValues(t, http.StatusNotFound, resp.StatusCode)
		assert.EqualValues(t, "no such user", result.Message)
	})

}
