package sendgrid

import (
	_ "embed"
	"errors"
	"fmt"
	"github.com/kotalco/core-api/config"
	"github.com/kotalco/core-api/core/setting"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/logger"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"net/http"
	"strings"
)

type service struct{}

type IService interface {
	SignUp(dto *MailRequestDto) restErrors.IRestErr
	ResendEmailVerification(dto *MailRequestDto) restErrors.IRestErr
	ForgetPassword(dto *MailRequestDto) restErrors.IRestErr
	WorkspaceInvitation(dto *WorkspaceInvitationMailRequestDto) restErrors.IRestErr
	Ping() restErrors.IRestErr
}

var (
	client    = GetClient()
	fromName  = config.Environment.SendgridSenderName
	fromEmail = config.Environment.SendgridsenderEmail
	greeting  = "Hello there!" //default value for user name
	//go:embed confirm_email.html
	ConfirmEmailTemplate string
	//go:embed reset_password.html
	ResetPasswordTemplate string
	//go:embed workspace_invitation.html
	WorkspaceInvitationTemplate string
)

func NewService() IService {
	newService := &service{}
	return newService
}

func (service) Ping() restErrors.IRestErr {
	if config.Environment.SendgridAPIKey == "" {
		restErr := restErrors.NewForbiddenError("Sendgrid is not configured")
		return restErr
	}
	request := sendgrid.GetRequest(config.Environment.SendgridAPIKey, "/v3/user/profile", "")
	request.Method = http.MethodGet

	response, err := sendgrid.API(request)
	if err != nil {
		return restErrors.NewForbiddenError(err.Error())
	} else {
		if response.StatusCode == http.StatusUnauthorized {
			return restErrors.NewForbiddenError("Sendgrid API key is not valid")
		}
		return nil
	}
}

func (service) SignUp(dto *MailRequestDto) restErrors.IRestErr {
	from := mail.NewEmail(fromName, fromEmail)
	subject := "Confirm Your Email"
	to := mail.NewEmail(greeting, dto.Email)
	plainTextContent := ""
	domainBaseUrl, restErr := setting.GetDomainBaseUrl()
	if restErr != nil {
		return restErr
	}
	baseUrl := fmt.Sprintf("https://app.%s/confirm-email?email=%s&token=%s", domainBaseUrl, dto.Email, dto.Token)
	content := strings.Replace(ConfirmEmailTemplate, "CALL_TO_ACTION_HREF", baseUrl, 1)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, content)

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

func (service) ResendEmailVerification(dto *MailRequestDto) restErrors.IRestErr {
	domainBaseUrl, restErr := setting.GetDomainBaseUrl()
	if restErr != nil {
		return restErr
	}
	from := mail.NewEmail(fromName, fromEmail)
	subject := "Confirm Your Email"
	to := mail.NewEmail(greeting, dto.Email)
	plainTextContent := ""
	baseUrl := fmt.Sprintf("https://app.%s/confirm-email?email=%s&token=%s", domainBaseUrl, dto.Email, dto.Token)
	content := strings.Replace(ConfirmEmailTemplate, "CALL_TO_ACTION_HREF", baseUrl, 1)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, content)

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

func (service) ForgetPassword(dto *MailRequestDto) restErrors.IRestErr {
	domainBaseUrl, restErr := setting.GetDomainBaseUrl()
	if restErr != nil {
		return restErr
	}
	from := mail.NewEmail(fromName, fromEmail)
	subject := "Reset Your Password"
	to := mail.NewEmail(greeting, dto.Email)
	plainTextContent := ""
	baseUrl := fmt.Sprintf("https://app.%s/reset-password?email=%s&token=%s", domainBaseUrl, dto.Email, dto.Token)
	content := strings.Replace(ResetPasswordTemplate, "CALL_TO_ACTION_HREF", baseUrl, 1)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, content)

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

func (service) WorkspaceInvitation(dto *WorkspaceInvitationMailRequestDto) restErrors.IRestErr {
	domainBaseUrl, restErr := setting.GetDomainBaseUrl()
	if restErr != nil {
		return restErr
	}
	from := mail.NewEmail(fromName, fromEmail)
	subject := "You're invited to join workspace"
	to := mail.NewEmail(greeting, dto.Email)
	plainTextContent := ""
	baseUrl := fmt.Sprintf("https://app.%s/workspaces/%s", domainBaseUrl, dto.WorkspaceId)
	content := strings.Replace(WorkspaceInvitationTemplate, "CALL_TO_ACTION_HREF", baseUrl, 1)
	content = strings.Replace(content, "KOTAL_WORKSPACE_NAME", dto.WorkspaceName, 1)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, content)

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
