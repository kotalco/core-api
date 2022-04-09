package sendgrid

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var sendGridTestService = NewService()

func TestService_SignUp(t *testing.T) {
	t.Run("SignUp_Email_Should_Pass", func(t *testing.T) {
		dto := new(MailRequestDto)
		dto.Token = "token"
		dto.Email = "m.abdelrhman@kotal.co"
		err := sendGridTestService.SignUp(dto)
		assert.Nil(t, err)
	})
}

func TestService_ForgetPassword(t *testing.T) {
	t.Run("ForgetPassword_Should_Pass", func(t *testing.T) {
		dto := new(MailRequestDto)
		dto.Token = "token"
		dto.Email = "m.abdelrhman@kotal.co"
		err := sendGridTestService.ForgetPassword(dto)
		assert.Nil(t, err)
	})
}

func TestService_ResendEmailVerification(t *testing.T) {
	t.Run("Test_Resend_EmailVerification_Should_Pass", func(t *testing.T) {
		dto := new(MailRequestDto)
		dto.Token = "token"
		dto.Email = "m.abdelrhman@kotal.co"
		err := sendGridTestService.ResendEmailVerification(dto)
		assert.Nil(t, err)
	})
}
