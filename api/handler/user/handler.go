package user

import (
	"database/sql"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kotalco/core-api/core/setting"
	"github.com/kotalco/core-api/core/user"
	"github.com/kotalco/core-api/core/verification"
	"github.com/kotalco/core-api/core/workspace"
	"github.com/kotalco/core-api/k8s"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/logger"
	"github.com/kotalco/core-api/pkg/responder"
	"github.com/kotalco/core-api/pkg/sendgrid"
	"github.com/kotalco/core-api/pkg/sqlclient"
	"github.com/kotalco/core-api/pkg/token"
	"net/http"
	"reflect"
	"strconv"
)

var (
	userService         = user.NewService()
	mailService         = sendgrid.NewService()
	verificationService = verification.NewService()
	workspaceService    = workspace.NewService()
	settingService      = setting.NewService()
	namespaceService    = k8s.NewNamespaceService()
)

// SignUp validate dto , create user , send verification token, create the default namespace and create the default workspace
func SignUp(c *fiber.Ctx) error {
	dto := new(user.SignUpRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}

	restErr := user.Validate(dto)
	if restErr != nil {
		return c.Status(restErr.StatusCode()).JSON(restErr)
	}

	if !settingService.WithoutTransaction().IsRegistrationEnabled() {
		err := restErrors.NewForbiddenError("Registration disabled")
		return c.Status(err.StatusCode()).JSON(err)
	}

	txRead := sqlclient.Begin(&sql.TxOptions{
		Isolation: sql.LevelReadUncommitted,
	})

	usersCount, restErr := userService.WithTransaction(txRead).Count()
	if restErr != nil {
		sqlclient.Rollback(txRead)
		return c.Status(restErr.StatusCode()).JSON(restErr)
	}
	sqlclient.Commit(txRead)

	txHandle := sqlclient.Begin()
	model, restErr := userService.WithTransaction(txHandle).SignUp(dto)
	if restErr != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(restErr.StatusCode()).JSON(restErr)
	}

	token, restErr := verificationService.WithTransaction(txHandle).Create(model.ID)
	if restErr != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(restErr.StatusCode()).JSON(restErr)
	}

	var defaultWorkspaceName = workspace.DefaultWorkspaceName
	var defaultNamespaceName = uuid.NewString()
	if reflect.ValueOf(usersCount).IsZero() { //check if this user is first user in the cluster=>verify email address
		restErr = verificationService.WithTransaction(txHandle).Verify(model.ID, token)
		if restErr != nil {
			sqlclient.Rollback(txHandle)
			return c.Status(restErr.StatusCode()).JSON(restErr)
		}
		restErr = userService.WithTransaction(txHandle).VerifyEmail(model)
		if restErr != nil {
			sqlclient.Rollback(txHandle)
			return c.Status(restErr.StatusCode()).JSON(restErr)
		}
		//set as platform admin
		restErr = userService.WithTransaction(txHandle).SetAsPlatformAdmin(model)
		if restErr != nil {
			sqlclient.Rollback(txHandle)
			return c.Status(restErr.StatusCode()).JSON(restErr)
		}
		// set registration to false
		enableRegistration := false
		restErr = settingService.WithTransaction(txHandle).ConfigureRegistration(&setting.ConfigureRegistrationRequestDto{
			EnableRegistration: &enableRegistration,
		})
		if restErr != nil {
			sqlclient.Rollback(txHandle)
			return c.Status(restErr.StatusCode()).JSON(restErr)
		}
		//since this is the first user in the cluster, it's default workspace should be bound with the default namespace
		defaultNamespaceName = "default"
	}

	//create the user default workspace
	_, restErr = workspaceService.WithTransaction(txHandle).Create(&workspace.CreateWorkspaceRequestDto{Name: defaultWorkspaceName}, model.ID, defaultNamespaceName)
	if restErr != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(restErr.StatusCode()).JSON(restErr)
	}
	//create the user default namespace
	//we shouldn't try to create the default namespace "default" it's always there
	if defaultNamespaceName != "default" {
		restErr = namespaceService.Create(defaultNamespaceName)
		if restErr != nil {
			sqlclient.Rollback(txHandle)
			return c.Status(restErr.StatusCode()).JSON(restErr)
		}
	}

	sqlclient.Commit(txHandle)

	//section that user don't need to wait for
	go func() {
		if usersCount > 1 { // if this user isn't the first user in the cluster send verification email
			//send email verification
			mailRequest := new(sendgrid.MailRequestDto)
			mailRequest.Token = token
			mailRequest.Email = model.Email
			mailService.SignUp(mailRequest)
		}
	}()

	return c.Status(http.StatusCreated).JSON(responder.NewResponse(new(user.UserResponseDto).Marshall(model)))
}

// SignIn creates bearer token for the yse
func SignIn(c *fiber.Ctx) error {

	dto := new(user.SignInRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}

	restErr := user.Validate(dto)
	if restErr != nil {
		return c.Status(restErr.StatusCode()).JSON(restErr)
	}

	session, restErr := userService.WithoutTransaction().SignIn(dto)
	if restErr != nil {
		return c.Status(restErr.StatusCode()).JSON(restErr)
	}

	return c.Status(http.StatusOK).JSON(responder.NewResponse(session))
}

// SendEmailVerification send email verification for user who
// users with email verification token got expired
// users who didn't receive email verification and want to resent token
func SendEmailVerification(c *fiber.Ctx) error {
	dto := new(user.SendEmailVerificationRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}

	err := user.Validate(dto)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	userModel, err := userService.WithoutTransaction().GetByEmail(dto.Email)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	if userModel.IsEmailVerified {
		badReq := restErrors.NewBadRequestError("email already verified")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}

	token, err := verificationService.WithoutTransaction().Resend(userModel.ID)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	mailRequest := new(sendgrid.MailRequestDto)
	mailRequest.Token = token
	mailRequest.Email = userModel.Email

	go mailService.ResendEmailVerification(mailRequest)

	//todo create shared successMessage struct in shared pkg
	resp := struct {
		Message string `json:"message"`
	}{
		Message: "email verification sent successfully",
	}

	return c.Status(http.StatusOK).JSON(responder.NewResponse(resp))
}

// VerifyEmail verify user email by email and token hash send to it's  email
func VerifyEmail(c *fiber.Ctx) error {
	dto := new(user.EmailVerificationRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}

	err := user.Validate(dto)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	userModel, err := userService.WithoutTransaction().GetByEmail(dto.Email)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	if userModel.IsEmailVerified {
		badReq := restErrors.NewBadRequestError("email already verified")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}

	txHandle := sqlclient.Begin()
	err = verificationService.WithTransaction(txHandle).Verify(userModel.ID, dto.Token)
	if err != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(err.StatusCode()).JSON(err)
	}

	err = userService.WithTransaction(txHandle).VerifyEmail(userModel)
	if err != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(err.StatusCode()).JSON(err)
	}

	sqlclient.Commit(txHandle)

	resp := struct {
		Message string `json:"message"`
	}{
		Message: "email verified",
	}

	return c.Status(http.StatusOK).JSON(responder.NewResponse(resp))
}

// ForgetPassword send verification token to user email to reset password
func ForgetPassword(c *fiber.Ctx) error {
	dto := new(user.SendEmailVerificationRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}

	err := user.Validate(dto)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	userModel, err := userService.WithoutTransaction().GetByEmail(dto.Email)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	token, err := verificationService.WithoutTransaction().Resend(userModel.ID)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	mailRequest := new(sendgrid.MailRequestDto)
	mailRequest.Token = token
	mailRequest.Email = userModel.Email

	go mailService.ForgetPassword(mailRequest)

	resp := struct {
		Message string `json:"message"`
	}{
		Message: "reset password has been sent to your email",
	}

	return c.Status(http.StatusOK).JSON(responder.NewResponse(resp))
}

// ResetPassword resets user password by accepting token hash and new password
func ResetPassword(c *fiber.Ctx) error {
	dto := new(user.ResetPasswordRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}

	err := user.Validate(dto)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	userModel, err := userService.WithoutTransaction().GetByEmail(dto.Email)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	if !userModel.IsEmailVerified {
		//todo change it to new forbidden once error deployed as package
		forbidErr := &restErrors.RestErr{
			Message: "email not verified",
			Status:  403,
			Name:    "Forbidden",
		}
		return c.Status(forbidErr.Status).JSON(forbidErr)
	}

	txHandle := sqlclient.Begin()

	err = verificationService.WithTransaction(txHandle).Verify(userModel.ID, dto.Token)
	if err != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(err.StatusCode()).JSON(err)
	}
	err = userService.WithTransaction(txHandle).ResetPassword(userModel, dto.Password)
	if err != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(err.StatusCode()).JSON(err)
	}

	sqlclient.Commit(txHandle)

	resp := struct {
		Message string `json:"message"`
	}{
		Message: "password reset successfully",
	}

	return c.Status(http.StatusOK).JSON(responder.NewResponse(resp))
}

// ChangePassword change user password
// todo log all user token out
func ChangePassword(c *fiber.Ctx) error {
	userId := c.Locals("user").(token.UserDetails).ID
	userDetails, err := userService.WithoutTransaction().GetById(userId)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	dto := new(user.ChangePasswordRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}

	err = user.Validate(dto)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	err = userService.WithoutTransaction().ChangePassword(userDetails, dto)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	//todo remove all user logins from redis authorization cache

	resp := struct {
		Message string `json:"message"`
	}{
		Message: "password changed successfully",
	}
	return c.Status(http.StatusOK).JSON(responder.NewResponse(resp))
}

// ChangeEmail change user email and send verification token to the user email
func ChangeEmail(c *fiber.Ctx) error {
	userId := c.Locals("user").(token.UserDetails).ID
	userDetails, err := userService.WithoutTransaction().GetById(userId)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	dto := new(user.ChangeEmailRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}

	err = user.Validate(dto)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	txHandle := sqlclient.Begin()

	err = userService.WithTransaction(txHandle).ChangeEmail(userDetails, dto)
	if err != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(err.StatusCode()).JSON(err)
	}

	token, err := verificationService.WithTransaction(txHandle).Resend(userDetails.ID)
	if err != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(err.StatusCode()).JSON(err)
	}

	sqlclient.Commit(txHandle)

	mailRequest := new(sendgrid.MailRequestDto)
	mailRequest.Token = token
	mailRequest.Email = userDetails.Email

	go mailService.ResendEmailVerification(mailRequest)

	resp := struct {
		Message string `json:"message"`
	}{
		Message: "email changed successfully",
	}
	return c.Status(http.StatusOK).JSON(responder.NewResponse(resp))
}

func Whoami(c *fiber.Ctx) error {
	userId := c.Locals("user").(token.UserDetails).ID
	userDetails, err := userService.WithoutTransaction().GetById(userId)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	return c.Status(http.StatusOK).JSON(responder.NewResponse(new(user.UserResponseDto).Marshall(userDetails)))

}

// CreateTOTP create time based one time password QR code so user can scan it with his mobile app
func CreateTOTP(c *fiber.Ctx) error {
	userId := c.Locals("user").(token.UserDetails).ID
	userDetails, err := userService.WithoutTransaction().GetById(userId)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	dto := new(user.CreateTOTPRequestDto)

	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}

	err = user.Validate(dto)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	qr, err := userService.WithoutTransaction().CreateTOTP(userDetails, dto)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	c.Set("Content-Type", "image/png")
	c.Set("Content-Length", strconv.Itoa(len(qr.Bytes())))
	if _, err := c.Write(qr.Bytes()); err != nil {
		go logger.Error(CreateTOTP, err)
		internalErr := restErrors.NewInternalServerError("some thing went wrong")

		return c.Status(internalErr.StatusCode()).JSON(internalErr)
	}
	return nil
}

// EnableTwoFactorAuth used one time when user scan the QR code to verify it scanned and configured correctly
// then it enables two-factor auth for the user
func EnableTwoFactorAuth(c *fiber.Ctx) error {
	userId := c.Locals("user").(token.UserDetails).ID
	userDetails, err := userService.WithoutTransaction().GetById(userId)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	dto := new(user.TOTPRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}

	model, err := userService.WithoutTransaction().EnableTwoFactorAuth(userDetails, dto.TOTP)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	return c.Status(http.StatusOK).JSON(responder.NewResponse(new(user.UserResponseDto).Marshall(model)))
}

// VerifyTOTP used after the login if the user enabled 2fa his bearer token will be limited to specific functions including this one
// create new bearer token for the user after totp validation
func VerifyTOTP(c *fiber.Ctx) error {
	userId := c.Locals("user").(token.UserDetails).ID
	userDetails, err := userService.WithoutTransaction().GetById(userId)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	dto := new(user.TOTPRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}

	session, err := userService.WithoutTransaction().VerifyTOTP(userDetails, dto.TOTP)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	return c.Status(http.StatusOK).JSON(responder.NewResponse(session))
}

func DisableTwoFactorAuth(c *fiber.Ctx) error {
	userId := c.Locals("user").(token.UserDetails).ID
	userDetails, err := userService.WithoutTransaction().GetById(userId)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	dto := new(user.DisableTOTPRequestDto)

	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}

	err = user.Validate(dto)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	err = userService.WithoutTransaction().DisableTwoFactorAuth(userDetails, dto)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	resp := struct {
		Message string `json:"message"`
	}{
		Message: "2FA disabled",
	}

	return c.Status(http.StatusOK).JSON(responder.NewResponse(resp))
}
