package user

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/api/pkg/shared"
	"github.com/kotalco/cloud-api/internal/user"
	"github.com/kotalco/cloud-api/internal/verification"
	"github.com/kotalco/cloud-api/pkg/sendgrid"
)

var (
	userService         = user.NewService()
	verificationService = verification.NewService()
)

//SignUp validate dto , create user , send verification token
func SignUp(c *fiber.Ctx) error {
	dto := new(user.SignUpRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.Status).JSON(badReq)
	}

	restErr := user.Validate(dto)
	if restErr != nil {
		return c.Status(restErr.Status).JSON(restErr)
	}

	model, restErr := userService.SignUp(dto)
	if restErr != nil {
		return c.Status(restErr.Status).JSON(restErr)
	}

	token, restErr := verificationService.Create(model.ID)
	if restErr != nil {
		return c.Status(restErr.Status).JSON(restErr)
	}

	//send email verification
	mailRequest := new(sendgrid.MailRequestDto)
	mailRequest.Token = token
	mailRequest.Email = model.Email

	go sendgrid.MailService.SignUp(*mailRequest)

	return c.Status(http.StatusCreated).JSON(shared.NewResponse(new(user.UserResponseDto).Marshall(model)))
}

//SignIn creates bearer token for the yse
func SignIn(c *fiber.Ctx) error {
	dto := new(user.SignInRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.Status).JSON(badReq)
	}

	restErr := user.Validate(dto)
	if restErr != nil {
		return c.Status(restErr.Status).JSON(restErr)
	}

	token, restErr := userService.SignIn(dto)
	if restErr != nil {
		return c.Status(restErr.Status).JSON(restErr)
	}

	session := new(user.UserSessionResponseDto)
	session.Token = token

	return c.Status(http.StatusOK).JSON(shared.NewResponse(session))
}

//SendEmailVerification send email verification for user who
//users with email verification token got expired
//users who didn't receive email verification and want to resent token
func SendEmailVerification(c *fiber.Ctx) error {
	dto := new(user.SendEmailVerificationRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.Status).JSON(badReq)
	}

	err := user.Validate(dto)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	userModel, err := userService.GetByEmail(dto.Email)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	if userModel.IsVerified {
		badReq := restErrors.NewBadRequestError("email already verified")
		return c.Status(badReq.Status).JSON(badReq)
	}

	token, err := verificationService.Resend(userModel.ID)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	mailRequest := new(sendgrid.MailRequestDto)
	mailRequest.Token = token
	mailRequest.Email = userModel.Email

	go sendgrid.MailService.ResendEmailVerification(*mailRequest)

	//todo create shared successMessage struct in shared pkg
	resp := struct {
		Message string `json:"message"`
	}{
		Message: "email verification sent successfully",
	}

	return c.Status(http.StatusCreated).JSON(shared.NewResponse(resp))
}

// VerifyEmail verify user email by email and token hash send to it's  email
func VerifyEmail(c *fiber.Ctx) error {
	dto := new(user.EmailVerificationRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.Status).JSON(badReq)
	}

	err := user.Validate(dto)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	userModel, err := userService.GetByEmail(dto.Email)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	err = verificationService.Verify(userModel.ID, dto.Token)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	err = userService.VerifyEmail(userModel)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	resp := struct {
		Message string `json:"message"`
	}{
		Message: "email verified",
	}

	return c.Status(http.StatusOK).JSON(shared.NewResponse(resp))
}

// ForgetPassword send verification token to user email to reset password
func ForgetPassword(c *fiber.Ctx) error {
	dto := new(user.SendEmailVerificationRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.Status).JSON(badReq)
	}

	err := user.Validate(dto)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	userModel, err := userService.GetByEmail(dto.Email)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	token, err := verificationService.Resend(userModel.ID)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	mailRequest := new(sendgrid.MailRequestDto)
	mailRequest.Token = token
	mailRequest.Email = userModel.Email

	go sendgrid.MailService.ForgetPassword(*mailRequest)

	resp := struct {
		Message string `json:"message"`
	}{
		Message: "reset password has been sent to your email",
	}

	return c.Status(http.StatusOK).JSON(shared.NewResponse(resp))
}

// ResetPassword resets user password by accepting token hash and new password
func ResetPassword(c *fiber.Ctx) error {
	dto := new(user.ResetPasswordRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.Status).JSON(badReq)
	}

	err := user.Validate(dto)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	userModel, err := userService.GetByEmail(dto.Email)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	if !userModel.IsVerified {
		//todo change it to new forbidden once error deployed as package
		forbidErr := &restErrors.RestErr{
			Message: "email not verified",
			Status:  403,
			Error:   "Forbidden",
		}
		return c.Status(forbidErr.Status).JSON(forbidErr)
	}

	err = verificationService.Verify(userModel.ID, dto.Token)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	err = userService.ResetPassword(userModel, dto.Password)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	resp := struct {
		Message string `json:"message"`
	}{
		Message: "password reset successfully",
	}

	return c.Status(http.StatusOK).JSON(shared.NewResponse(resp))
}

//ChangePassword change user password
//todo log all user tokens out
func ChangePassword(c *fiber.Ctx) error {
	authorizedUser := c.Locals("user").(*user.User)
	dto := new(user.ChangePasswordRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.Status).JSON(badReq)
	}

	err := user.Validate(dto)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	err = userService.ChangePassword(authorizedUser, dto)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	//todo remove all user logins from redis authorization cache

	resp := struct {
		Message string `json:"message"`
	}{
		Message: "password changed successfully",
	}
	return c.Status(http.StatusOK).JSON(shared.NewResponse(resp))
}

//ChangeEmail change user email and send verification token to the user email
func ChangeEmail(c *fiber.Ctx) error {
	authorizedUser := c.Locals("user").(*user.User)
	dto := new(user.ChangeEmailRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.Status).JSON(badReq)
	}

	err := user.Validate(dto)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	err = userService.ChangeEmail(authorizedUser, dto)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	token, err := verificationService.Resend(authorizedUser.ID)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	mailRequest := new(sendgrid.MailRequestDto)
	mailRequest.Token = token
	mailRequest.Email = authorizedUser.Email

	go sendgrid.MailService.ResendEmailVerification(*mailRequest)

	resp := struct {
		Message string `json:"message"`
	}{
		Message: "email changed successfully",
	}
	return c.Status(http.StatusOK).JSON(shared.NewResponse(resp))
}
