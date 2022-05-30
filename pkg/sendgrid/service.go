package sendgrid

import (
	"errors"
	"fmt"

	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/cloud-api/pkg/config"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type service struct{}

type IService interface {
	SignUp(dto *MailRequestDto) *restErrors.RestErr
	ResendEmailVerification(dto *MailRequestDto) *restErrors.RestErr
	ForgetPassword(dto *MailRequestDto) *restErrors.RestErr
	WorkspaceInvitation(dto *WorkspaceInvitationMailRequestDto) *restErrors.RestErr
}

var (
	client    = GetClient()
	fromName  = config.EnvironmentConf["SEND_GRID_SENDER_NAME"]
	fromEmail = config.EnvironmentConf["SEND_GRID_SENDER_EMAIL"]
	greeting  = "Hello there!" //default value for user name
)

func NewService() IService {
	newService := &service{}
	return newService
}

func (service) SignUp(dto *MailRequestDto) *restErrors.RestErr {
	from := mail.NewEmail(fromName, fromEmail)
	subject := "Welcome to Kotal! Confirm Your Email"
	to := mail.NewEmail(greeting, dto.Email)
	plainTextContent := ""
	baseUrl := fmt.Sprintf("%s/confirm-email?email=%s&token=%s", config.EnvironmentConf["FRONT_BASE_URL"], dto.Email, dto.Token)
	htmlContent := fmt.Sprintf("please visit the following link to Confirm  your email address %s", baseUrl)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)

	restResponse, err := client.Send(message)
	if err != nil {
		go logger.Error(service.SignUp, err)
		return restErrors.NewInternalServerError("some thing went wrong")
	}
	if restResponse.StatusCode >= 400 {
		go logger.Error(service.SignUp, errors.New(restResponse.Body))
		return restErrors.NewInternalServerError("some thing went wrong")
	}

	return nil
}

func (service) ResendEmailVerification(dto *MailRequestDto) *restErrors.RestErr {
	from := mail.NewEmail(fromName, fromEmail)
	subject := "Confirm Your Email"
	to := mail.NewEmail(greeting, dto.Email)
	plainTextContent := ""
	baseUrl := fmt.Sprintf("%s/confirm-email?email=%s&token=%s", config.EnvironmentConf["FRONT_BASE_URL"], dto.Email, dto.Token)
	htmlContent := fmt.Sprintf("please visit the following link to Confirm  your email address %s", baseUrl)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)

	restResponse, err := client.Send(message)
	if err != nil {
		go logger.Error(service.SignUp, err)
		return restErrors.NewInternalServerError("some thing went wrong")
	}
	if restResponse.StatusCode >= 400 {
		go logger.Error(service.SignUp, errors.New(restResponse.Body))
		return restErrors.NewInternalServerError("some thing went wrong")
	}

	return nil
}

func (service) ForgetPassword(dto *MailRequestDto) *restErrors.RestErr {
	from := mail.NewEmail(fromName, fromEmail)
	subject := "Reset Password"
	to := mail.NewEmail(greeting, dto.Email)
	plainTextContent := ""
	baseUrl := fmt.Sprintf("%s/reset-password?email=%s&token=%s", config.EnvironmentConf["FRONT_BASE_URL"], dto.Email, dto.Token)
	htmlContent := fmt.Sprintf("please visit the following link to reset  your password %s", baseUrl)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)

	restResponse, err := client.Send(message)
	if err != nil {
		go logger.Error(service.SignUp, err)
		return restErrors.NewInternalServerError("some thing went wrong")
	}
	if restResponse.StatusCode >= 400 {
		go logger.Error(service.SignUp, errors.New(restResponse.Body))
		return restErrors.NewInternalServerError("some thing went wrong")
	}

	return nil
}

func (service) WorkspaceInvitation(dto *WorkspaceInvitationMailRequestDto) *restErrors.RestErr {
	from := mail.NewEmail(fromName, fromEmail)
	subject := "Workspace Invitation"
	to := mail.NewEmail(greeting, dto.Email)
	plainTextContent := ""
	baseUrl := fmt.Sprintf("%s/workspaces/%s", config.EnvironmentConf["FRONT_BASE_URL"], dto.WorkspaceId)
	htmlContent := fmt.Sprintf("You've been invited to %s workspace..   %s ", dto.WorkspaceName, baseUrl)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)

	restResponse, err := client.Send(message)
	if err != nil {
		go logger.Error(service.WorkspaceInvitation, err)
		return restErrors.NewInternalServerError("some thing went wrong")
	}
	if restResponse.StatusCode >= 400 {
		go logger.Error(service.WorkspaceInvitation, errors.New(restResponse.Body))
		return restErrors.NewInternalServerError("some thing went wrong")
	}

	return nil
}
