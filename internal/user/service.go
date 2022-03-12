package user

import (
	"github.com/google/uuid"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/cloud-api/pkg/security"
	"github.com/kotalco/cloud-api/pkg/tokens"
)

type service struct{}

type IService interface {
	SignUp(dto *SignUpRequestDto) (*User, *restErrors.RestErr)
	SignIn(dto *SignInRequestDto) (string, *restErrors.RestErr)
	GetByEmail(email string) (*User, *restErrors.RestErr)
	VerifyEmail(model *User) *restErrors.RestErr
	ResetPassword(model *User, password string) *restErrors.RestErr
	ChangePassword(model *User, dto *ChangePasswordRequestDto) *restErrors.RestErr
	ChangeEmail(model *User, dto *ChangeEmailRequestDto) *restErrors.RestErr
}

var (
	userRepository = NewRepository()
)

func NewService() IService {
	newService := &service{}
	return newService
}

//SignUp Creates new user
func (service) SignUp(dto *SignUpRequestDto) (*User, *restErrors.RestErr) {
	hashedPassword, err := security.Hash(dto.Password, 13)
	if err != nil {
		go logger.Error(service.SignUp, err)
		return nil, restErrors.NewInternalServerError("something went wrong.")
	}

	user := new(User)
	user.ID = uuid.New().String()
	user.Email = dto.Email
	user.IsVerified = false
	user.Password = string(hashedPassword)

	restErr := userRepository.Create(user)
	if restErr != nil {
		return nil, restErr
	}

	return user, nil
}

//SignIn Log user in and  returns jwt token
func (service) SignIn(dto *SignInRequestDto) (string, *restErrors.RestErr) {
	user, err := userRepository.GetByEmail(dto.Email)
	if err != nil {
		return "", restErrors.NewUnAuthorizedError("Invalid Credentials")
	}

	if !user.IsVerified {
		//todo change it to new forbidden once error deployed as package
		return "", &restErrors.RestErr{
			Message: "email not verified",
			Status:  403,
			Error:   "Forbidden",
		}
	}

	invalidPassError := security.VerifyPassword(user.Password, dto.Password)
	if invalidPassError != nil {
		return "", restErrors.NewUnAuthorizedError("Invalid Credentials")
	}

	token, err := tokens.CreateToken(user.ID, dto.RememberMe)
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

//GetByEmail find user by email
func (service) GetByEmail(email string) (*User, *restErrors.RestErr) {
	model, err := userRepository.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	return model, nil
}

//VerifyEmail change user isVerified to true
//user can't sign in if this field is falsy
func (service) VerifyEmail(model *User) *restErrors.RestErr {
	model.IsVerified = true
	err := userRepository.Update(model)
	if err != nil {
		return err
	}
	return nil
}

//ResetPassword create new password for user used after user ForgetPassword
func (service) ResetPassword(model *User, password string) *restErrors.RestErr {
	hashedPassword, error := security.Hash(password, 13)
	if error != nil {
		go logger.Error(service.ResetPassword, error)
		return restErrors.NewInternalServerError("something went wrong.")
	}

	model.Password = string(hashedPassword)

	err := userRepository.Update(model)
	if err != nil {
		return err
	}

	return nil
}

//ChangePassword change user password  for authenticated users
func (service) ChangePassword(model *User, dto *ChangePasswordRequestDto) *restErrors.RestErr {

	invalidPassError := security.VerifyPassword(model.Password, dto.OldPassword)
	if invalidPassError != nil {
		return restErrors.NewUnAuthorizedError("invalid old password")
	}

	hashedPassword, error := security.Hash(dto.Password, 13)
	if error != nil {
		go logger.Error(service.ChangePassword, error)
		return restErrors.NewInternalServerError("something went wrong.")
	}

	model.Password = string(hashedPassword)

	err := userRepository.Update(model)
	if err != nil {
		return err
	}

	return nil
}

//ChangeEmail change user email for authenticated users
func (service) ChangeEmail(model *User, dto *ChangeEmailRequestDto) *restErrors.RestErr {
	model.Email = dto.Email
	model.IsVerified = false

	err := userRepository.Update(model)
	if err != nil {
		return err
	}

	return nil
}
