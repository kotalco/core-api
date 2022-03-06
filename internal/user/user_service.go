package user

import (
	"github.com/google/uuid"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/cloud-api/pkg/security"
	"github.com/kotalco/cloud-api/pkg/tokens"
)

type userService struct{}

type userServiceInterface interface {
	SignUp(dto SignUpRequestDto) (*User, *restErrors.RestErr)
	SignIn(dto SignInRequestDto) (*string, *restErrors.RestErr)
	GetByEmail(email string) (*User, *restErrors.RestErr)
	VerifyEmail(model *User) *restErrors.RestErr
	ResetPassword(model *User, password string) *restErrors.RestErr
	ChangePassword(model *User, dto ChangePasswordRequestDto) *restErrors.RestErr
	ChangeEmail(model *User, dto ChangeEmailRequestDto) *restErrors.RestErr
}

var (
	UserService userServiceInterface
)

func init() { UserService = &userService{} }

//SignUp Creates new user
func (service userService) SignUp(dto SignUpRequestDto) (*User, *restErrors.RestErr) {
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

	restErr := UserRepository.Create(user)
	if restErr != nil {
		return nil, restErr
	}

	return user, nil
}

//SignIn Log user in and  returns jwt token
func (service userService) SignIn(dto SignInRequestDto) (*string, *restErrors.RestErr) {
	user, err := UserRepository.GetByEmail(dto.Email)
	if err != nil {
		return nil, restErrors.NewUnAuthorizedError("Invalid Credentials")
	}

	if !user.IsVerified {
		//todo change it to new forbidden once error deployed as package
		return nil, &restErrors.RestErr{
			Message: "email not verified",
			Status:  403,
			Error:   "Forbidden",
		}
	}

	invalidPassError := security.VerifyPassword(user.Password, dto.Password)
	if invalidPassError != nil {
		return nil, restErrors.NewUnAuthorizedError("Invalid Credentials")
	}

	token, err := tokens.CreateToken(user.ID, dto.RememberMe)
	if err != nil {
		return nil, err
	}

	return &token.AccessToken, nil
}

//GetByEmail find user by email
func (service userService) GetByEmail(email string) (*User, *restErrors.RestErr) {
	model, err := UserRepository.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	return model, nil
}

//VerifyEmail change user isVerified to true
//user can't sign in if this field is falsy
func (service userService) VerifyEmail(model *User) *restErrors.RestErr {
	model.IsVerified = true
	err := UserRepository.Update(model)
	if err != nil {
		return err
	}
	return nil
}

//ResetPassword create new password for user used after user ForgetPassword
func (service userService) ResetPassword(model *User, password string) *restErrors.RestErr {
	hashedPassword, error := security.Hash(password, 13)
	if error != nil {
		go logger.Error(service.ResetPassword, error)
		return restErrors.NewInternalServerError("something went wrong.")
	}

	model.Password = string(hashedPassword)

	err := UserRepository.Update(model)
	if err != nil {
		return err
	}

	return nil
}

//ChangePassword change user password  for authenticated users
func (service userService) ChangePassword(model *User, dto ChangePasswordRequestDto) *restErrors.RestErr {

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

	err := UserRepository.Update(model)
	if err != nil {
		return err
	}

	return nil
}

//ChangeEmail change user email for authenticated users
func (service userService) ChangeEmail(model *User, dto ChangeEmailRequestDto) *restErrors.RestErr {
	model.Email = dto.Email
	model.IsVerified = false

	err := UserRepository.Update(model)
	if err != nil {
		return err
	}

	return nil
}
