package user

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/api/pkg/shared"
	"github.com/kotalco/cloud-api/internal/user"
	"github.com/kotalco/cloud-api/internal/verification"
	"github.com/kotalco/cloud-api/pkg/sendgrid"
)

var (
	userService         = user.NewService()
	mailService         = sendgrid.NewService()
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

	go mailService.SignUp(*mailRequest)

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

	session, restErr := userService.SignIn(*dto)
	if restErr != nil {
		return c.Status(restErr.Status).JSON(restErr)
	}

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

	if userModel.IsEmailVerified {
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

	go mailService.ResendEmailVerification(*mailRequest)

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

	go mailService.ForgetPassword(*mailRequest)

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

	if !userModel.IsEmailVerified {
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

	go mailService.ResendEmailVerification(*mailRequest)

	resp := struct {
		Message string `json:"message"`
	}{
		Message: "email changed successfully",
	}
	return c.Status(http.StatusOK).JSON(shared.NewResponse(resp))
}

//CreateTOTP create time based one time password QR code so user can scan it with his mobile app
func CreateTOTP(c *fiber.Ctx) error {
	authorizedUser := c.Locals("user").(*user.User)

	qr, err := userService.CreateTOTP(authorizedUser)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	c.Set("Content-Type", "image/png")
	c.Set("Content-Length", strconv.Itoa(len(qr.Bytes())))
	if _, err := c.Write(qr.Bytes()); err != nil {
		log.Println("unable to write image.")
	}
	return nil
}

//EnableTwoFactorAuth used one time when user scan the QR code to verify it scanned and configured correctly
//then it enables two-factor auth for the user
func EnableTwoFactorAuth(c *fiber.Ctx) error {
	authorizedUser := c.Locals("user").(*user.User)

	dto := new(user.TOTPRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.Status).JSON(badReq)
	}

	model, err := userService.EnableTwoFactorAuth(authorizedUser, dto.TOTP)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	return c.Status(http.StatusOK).JSON(shared.NewResponse(new(user.UserResponseDto).Marshall(model)))
}

//VerifyTOTP used after the login if the user enabled 2fa his bearer token will be limited to specific functions including this one
//create new bearer token for the user after totp validation
func VerifyTOTP(c *fiber.Ctx) error {
	authorizedUser := c.Locals("user").(*user.User)

	dto := new(user.TOTPRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.Status).JSON(badReq)
	}

	session, err := userService.VerifyTOTP(authorizedUser, dto.TOTP)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	return c.Status(http.StatusOK).JSON(shared.NewResponse(session))
}

func DisableTwoFactorAuth(c *fiber.Ctx) error {
	authorizedUser := c.Locals("user").(*user.User)

	err := userService.DisableTwoFactorAuth(authorizedUser)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	resp := struct {
		Message string `json:"message"`
	}{
		Message: "2FA disabled",
	}

	return c.Status(http.StatusOK).JSON(shared.NewResponse(resp))

}
