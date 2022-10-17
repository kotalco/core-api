package user

import (
	"github.com/go-playground/validator/v10"
	restErrors "github.com/kotalco/community-api/pkg/errors"
)

type SignUpRequestDto struct {
	Email                string `json:"email" validate:"required,email,lte=100"`
	Password             string `json:"password" validate:"gte=6,lte=100"`
	PasswordConfirmation string `json:"password_confirmation" validate:"required,lte=255,eqcsfield=Password"`
}

type SignInRequestDto struct {
	Email      string `json:"email" validate:"required,email,lte=100"`
	Password   string `json:"password" validate:"required,gte=6,lte=255"`
	RememberMe bool   `json:"remember_me"`
}

type ResetPasswordRequestDto struct {
	Email                string `json:"email" validate:"required,email,lte=100"`
	Password             string `json:"password" validate:"gte=6,lte=100"`
	PasswordConfirmation string `json:"password_confirmation" validate:"required,lte=255,eqcsfield=Password"`
	Token                string `json:"token" validate:"required,len=80"` //length of sent Token defined in env_conf
}

type ChangePasswordRequestDto struct {
	OldPassword          string `json:"old_password" validate:"gte=6,lte=100"`
	Password             string `json:"password" validate:"gte=6,lte=100"`
	PasswordConfirmation string `json:"password_confirmation" validate:"required,lte=255,eqcsfield=Password"`
}

type SendEmailVerificationRequestDto struct {
	Email string `json:"email" validate:"required,email,lte=100"`
}

type EmailVerificationRequestDto struct {
	Email string `json:"email" validate:"required,email,lte=100"`
	Token string `json:"token" validate:"required,len=80"` //length of sent Token defined in env_conf
}

type ChangeEmailRequestDto struct {
	Email    string `json:"email" validate:"required,email,lte=100"`
	Password string `json:"password" validate:"required,gte=6,lte=100"`
}

type CreateTOTPRequestDto struct {
	Password string `json:"password" validate:"required,gte=6,lte=100"`
}

type DisableTOTPRequestDto struct {
	Password string `json:"password" validate:"required,gte=6,lte=100"`
}

type TOTPRequestDto struct {
	TOTP string `json:"totp"`
}

type UserResponseDto struct {
	PublicUserResponseDto
	TwoFactorEnabled bool `json:"two_factor_enabled"`
}

type PublicUserResponseDto struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role,omitempty"`
}

type UserSessionResponseDto struct {
	Authorized bool
	Token      string `json:"token"`
}

// Marshall creates user response from user model
func (dto UserResponseDto) Marshall(model *User) UserResponseDto {
	dto.ID = model.ID
	dto.Email = model.Email
	dto.TwoFactorEnabled = model.TwoFactorEnabled
	return dto
}

// Marshall creates user response from user model
func (dto PublicUserResponseDto) Marshall(model *User) PublicUserResponseDto {
	dto.ID = model.ID
	dto.Email = model.Email
	return dto
}

// Validate validates user request field for all request dto`s
func Validate(dto interface{}) *restErrors.RestErr {
	newValidator := validator.New()
	err := newValidator.Struct(dto)

	if err != nil {
		fields := map[string]string{}
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Field() {
			case "Email":
				fields["email"] = "invalid email address"
				break
			case "Password":
				fields["password"] = "password should be at least 6 chars"
				break
			case "PasswordConfirmation":
				fields["password_confirmation"] = "password_confirmation should match password field"
				break
			case "Token":
				fields["token"] = "invalid token signature"
				break
			case "OldPassword":
				fields["old_password"] = "password should be at least 6 chars"
				break
			}
		}

		if len(fields) > 0 {
			return restErrors.NewValidationError(fields)
		}
	}

	return nil
}
