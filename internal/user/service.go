package user

import (
	"bytes"

	"github.com/google/uuid"
	"github.com/kotalco/cloud-api/pkg/config"
	"github.com/kotalco/cloud-api/pkg/security"
	"github.com/kotalco/cloud-api/pkg/tfa"
	"github.com/kotalco/cloud-api/pkg/token"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	"gorm.io/gorm"
)

type service struct{}

type IService interface {
	WithTransaction(txHandle gorm.DB) IService
	SignUp(dto *SignUpRequestDto) (*User, *restErrors.RestErr)
	SignIn(dto *SignInRequestDto) (*UserSessionResponseDto, *restErrors.RestErr)
	VerifyTOTP(model *User, totp string) (*UserSessionResponseDto, *restErrors.RestErr)
	GetByEmail(email string) (*User, *restErrors.RestErr)
	GetById(ID string) (*User, *restErrors.RestErr)
	VerifyEmail(model *User) *restErrors.RestErr
	ResetPassword(model *User, password string) *restErrors.RestErr
	ChangePassword(model *User, dto *ChangePasswordRequestDto) *restErrors.RestErr
	ChangeEmail(model *User, dto *ChangeEmailRequestDto) *restErrors.RestErr
	CreateTOTP(model *User, dto *CreateTOTPRequestDto) (bytes.Buffer, *restErrors.RestErr)
	EnableTwoFactorAuth(model *User, totp string) (*User, *restErrors.RestErr)
	DisableTwoFactorAuth(model *User, dto *DisableTOTPRequestDto) *restErrors.RestErr
	FindWhereIdInSlice(ids []string) ([]*User, *restErrors.RestErr)
	Count() (int64, *restErrors.RestErr)
}

var (
	userRepository = NewRepository()
	encryption     = security.NewEncryption()
	hashing        = security.NewHashing()
	tfaService     = tfa.NewTfa()
	tokenService   = token.NewToken()
)

func NewService() IService {
	newService := &service{}
	return newService
}

func (uService service) WithTransaction(txHandle gorm.DB) IService {
	userRepository = userRepository.WithTransaction(txHandle)
	return uService
}

// SignUp Creates new user
func (service) SignUp(dto *SignUpRequestDto) (*User, *restErrors.RestErr) {
	hashedPassword, err := hashing.Hash(dto.Password, 13)
	if err != nil {
		go logger.Error(service.SignUp, err)
		return nil, restErrors.NewInternalServerError("something went wrong.")
	}

	user := new(User)
	user.ID = uuid.New().String()
	user.Email = dto.Email
	user.IsEmailVerified = false
	user.Password = string(hashedPassword)

	restErr := userRepository.Create(user)
	if restErr != nil {
		return nil, restErr
	}

	return user, nil
}

// SignIn Log user in and  returns jwt token
func (service) SignIn(dto *SignInRequestDto) (*UserSessionResponseDto, *restErrors.RestErr) {
	user, err := userRepository.GetByEmail(dto.Email)
	if err != nil {
		return nil, restErrors.NewUnAuthorizedError("Invalid Credentials")
	}

	if !user.IsEmailVerified {
		//todo change it to new forbidden once error deployed as package
		return nil, &restErrors.RestErr{
			Message: "email not verified",
			Status:  403,
			Name:    "Forbidden",
		}
	}

	invalidPassError := hashing.VerifyHash(user.Password, dto.Password)
	if invalidPassError != nil {
		return nil, restErrors.NewUnAuthorizedError("Invalid Credentials")
	}

	var authorized bool
	if user.TwoFactorEnabled {
		authorized = false
	} else {
		authorized = true
	}

	token, err := tokenService.CreateToken(user.ID, dto.RememberMe, authorized)
	if err != nil {
		return nil, err
	}

	session := new(UserSessionResponseDto)
	session.Token = token.AccessToken
	session.Authorized = authorized

	return session, nil
}

// GetByEmail find user by email
func (service) GetByEmail(email string) (*User, *restErrors.RestErr) {
	model, err := userRepository.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	return model, nil
}

// GetById get user by Id
func (service) GetById(ID string) (*User, *restErrors.RestErr) {
	model, err := userRepository.GetById(ID)
	if err != nil {
		return nil, err
	}

	return model, nil
}

// VerifyEmail change user isVerified to true
// user can't sign in if this field is falsy
func (service) VerifyEmail(model *User) *restErrors.RestErr {
	model.IsEmailVerified = true
	err := userRepository.Update(model)
	if err != nil {
		return err
	}
	return nil
}

// ResetPassword create new password for user used after user ForgetPassword
func (service) ResetPassword(model *User, password string) *restErrors.RestErr {
	hashedPassword, error := hashing.Hash(password, 13)
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

// ChangePassword change user password  for authenticated users
func (service) ChangePassword(model *User, dto *ChangePasswordRequestDto) *restErrors.RestErr {

	invalidPassError := hashing.VerifyHash(model.Password, dto.OldPassword)
	if invalidPassError != nil {
		return restErrors.NewUnAuthorizedError("invalid old password")
	}

	hashedPassword, error := hashing.Hash(dto.Password, 13)
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

// ChangeEmail change user email for authenticated users
func (service) ChangeEmail(model *User, dto *ChangeEmailRequestDto) *restErrors.RestErr {

	invalidPassError := hashing.VerifyHash(model.Password, dto.Password)
	if invalidPassError != nil {
		return restErrors.NewBadRequestError("invalid password")
	}

	model.Email = dto.Email
	model.IsEmailVerified = false

	err := userRepository.Update(model)
	if err != nil {
		return err
	}

	return nil
}

// CreateTOTP Enables two-factor auth for users using qr-code
// if the user requested another qr code he/she has to register the new one on the 2auth-app coz the old one became invalid
func (service) CreateTOTP(model *User, dto *CreateTOTPRequestDto) (bytes.Buffer, *restErrors.RestErr) {

	invalidPassError := hashing.VerifyHash(model.Password, dto.Password)
	if invalidPassError != nil {
		return bytes.Buffer{}, restErrors.NewBadRequestError("Invalid password")
	}

	qrBytes, secret, err := tfaService.CreateQRCode(model.Email)
	if err != nil {
		go logger.Error(service.CreateTOTP, err)
		return bytes.Buffer{}, restErrors.NewInternalServerError("something went wrong")
	}

	twoAuthSecretCipher, err := encryption.Encrypt([]byte(secret), config.Environment.TwoFactorSecret)
	if err != nil {
		go logger.Error(service.CreateTOTP, err)
		return bytes.Buffer{}, restErrors.NewInternalServerError("something went wrong")
	}

	model.TwoFactorCipher = twoAuthSecretCipher
	restErr := userRepository.Update(model)
	if restErr != nil {
		return bytes.Buffer{}, restErr
	}

	return qrBytes, nil
}

// EnableTwoFactorAuth enables two-factor auth for the user after checking first time otp is valid
func (service) EnableTwoFactorAuth(model *User, totp string) (*User, *restErrors.RestErr) {
	if model.TwoFactorCipher == "" {
		return nil, restErrors.NewBadRequestError("please create and register qr code first")
	}
	TOTPSecret, err := encryption.Decrypt(model.TwoFactorCipher, config.Environment.TwoFactorSecret)
	if err != nil {
		go logger.Error(service.EnableTwoFactorAuth, err)
		return nil, restErrors.NewInternalServerError("something went wrong")
	}

	valid := tfaService.CheckOtp(TOTPSecret, totp)
	if !valid {
		return nil, restErrors.NewBadRequestError("invalid totp code")
	}

	model.TwoFactorEnabled = true
	restErr := userRepository.Update(model)
	if restErr != nil {
		return nil, restErr
	}

	return model, nil
}

// VerifyTOTP used after SignIn if the user enabled the 2fa to create another bearer token
func (service) VerifyTOTP(model *User, totp string) (*UserSessionResponseDto, *restErrors.RestErr) {
	if !model.TwoFactorEnabled {
		return nil, restErrors.NewBadRequestError("please enable your 2fa first")
	}

	TOTPSecret, err := encryption.Decrypt(model.TwoFactorCipher, config.Environment.TwoFactorSecret)
	if err != nil {
		go logger.Error(service.EnableTwoFactorAuth, err)
		return nil, restErrors.NewInternalServerError("something went wrong")
	}

	valid := tfaService.CheckOtp(TOTPSecret, totp)
	if !valid {
		return nil, restErrors.NewBadRequestError("invalid totp code")
	}

	token, restErr := tokenService.CreateToken(model.ID, true, true)

	if restErr != nil {
		return nil, restErr
	}

	session := new(UserSessionResponseDto)
	session.Token = token.AccessToken
	session.Authorized = true

	return session, nil
}

// DisableTwoFactorAuth disables two-factor auth for the user
func (service) DisableTwoFactorAuth(model *User, dto *DisableTOTPRequestDto) *restErrors.RestErr {
	invalidPassError := hashing.VerifyHash(model.Password, dto.Password)
	if invalidPassError != nil {
		return restErrors.NewBadRequestError("Invalid password")
	}

	if !model.TwoFactorEnabled {
		return restErrors.NewBadRequestError("2fa already disabled")
	}
	model.TwoFactorEnabled = false
	model.TwoFactorCipher = ""
	err := userRepository.Update(model)

	if err != nil {
		return err
	}

	return nil
}

// FindWhereIdInSlice returns a list of users which ids exist in the slice of ids passed as argument
func (service) FindWhereIdInSlice(ids []string) ([]*User, *restErrors.RestErr) {
	return userRepository.FindWhereIdInSlice(ids)
}

// Count returns count of the users model
func (service) Count() (int64, *restErrors.RestErr) {
	return userRepository.Count()
}
