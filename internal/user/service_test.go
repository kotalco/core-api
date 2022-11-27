package user

import (
	"bytes"
	"errors"
	"github.com/kotalco/cloud-api/pkg/token"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"net/http"
	"os"
	"testing"
)

var (
	WithTransactionFunc    func(txHandle *gorm.DB) IRepository
	userService            IService
	CreateFunc             func(user *User) *restErrors.RestErr
	GetByEmailFunc         func(email string) (*User, *restErrors.RestErr)
	GetByIdFunc            func(id string) (*User, *restErrors.RestErr)
	UpdateFunc             func(user *User) *restErrors.RestErr
	FindWhereIdInSliceFunc func(ids []string) ([]*User, *restErrors.RestErr)
	CountFunc              func() (int64, *restErrors.RestErr)

	EncryptFunc func(data []byte, passphrase string) (string, error)
	DecryptFunc func(encodedCipher string, passphrase string) (string, error)

	HashFunc       func(password string, cost int) ([]byte, error)
	VerifyHashFunc func(hashedPassword, password string) error

	CreateTokenFunc          func(userId string, rememberMe bool, authorized bool) (*token.Token, *restErrors.RestErr)
	ExtractTokenMetadataFunc func(bearToken string) (*token.AccessDetails, *restErrors.RestErr)

	CreateQRCodeFunc func(accountName string) (bytes.Buffer, string, error)
	CheckOtpFunc     func(userTOTPSecret string, otp string) bool
)

type userRepositoryMock struct{}
type encryptionServiceMock struct{}
type hashingServiceMock struct{}
type tokenServiceMock struct{}
type tfaServiceMock struct{}

// user repository methods
func (r userRepositoryMock) WithTransaction(txHandle *gorm.DB) IRepository {
	return r
}

func (userRepositoryMock) Create(user *User) *restErrors.RestErr {
	return CreateFunc(user)
}

func (userRepositoryMock) GetByEmail(email string) (*User, *restErrors.RestErr) {
	return GetByEmailFunc(email)
}

func (userRepositoryMock) GetById(id string) (*User, *restErrors.RestErr) {
	return GetByIdFunc(id)
}

func (userRepositoryMock) Update(user *User) *restErrors.RestErr {
	return UpdateFunc(user)
}

func (userRepositoryMock) FindWhereIdInSlice(ids []string) ([]*User, *restErrors.RestErr) {
	return FindWhereIdInSliceFunc(ids)
}
func (userRepositoryMock) Count() (int64, *restErrors.RestErr) {
	return CountFunc()
}

// encryption service methods
func (encryptionServiceMock) Encrypt(data []byte, passphrase string) (string, error) {
	return EncryptFunc(data, passphrase)
}

func (encryptionServiceMock) Decrypt(encodedCipher string, passphrase string) (string, error) {
	return DecryptFunc(encodedCipher, passphrase)
}

// hashing service methods
func (hashingServiceMock) Hash(password string, cost int) ([]byte, error) {
	return HashFunc(password, cost)
}

func (hashingServiceMock) VerifyHash(hashedPassword, password string) error {
	return VerifyHashFunc(hashedPassword, password)
}

// token service methods
func (tokenServiceMock) CreateToken(userId string, rememberMe bool, authorized bool) (*token.Token, *restErrors.RestErr) {
	return CreateTokenFunc(userId, rememberMe, authorized)
}

func (tokenServiceMock) ExtractTokenMetadata(bearToken string) (*token.AccessDetails, *restErrors.RestErr) {
	return ExtractTokenMetadataFunc(bearToken)
}

// tfa service methods
func (tfaServiceMock) CreateQRCode(accountName string) (bytes.Buffer, string, error) {
	return CreateQRCodeFunc(accountName)
}

func (tfaServiceMock) CheckOtp(userTOTPSecret string, otp string) bool {
	return CheckOtpFunc(userTOTPSecret, otp)
}

func TestMain(m *testing.M) {
	userRepository = &userRepositoryMock{}
	encryption = &encryptionServiceMock{}
	hashing = &hashingServiceMock{}
	tokenService = &tokenServiceMock{}
	tfaService = &tfaServiceMock{}

	userService = NewService()
	code := m.Run()
	os.Exit(code)
}

func TestService_SignUp(t *testing.T) {
	dto := new(SignUpRequestDto)
	dto.Password = "123456"

	t.Run("Sign_Up_Should_Pass", func(t *testing.T) {
		HashFunc = func(password string, cost int) ([]byte, error) {
			return []byte(""), nil
		}

		CreateFunc = func(user *User) *restErrors.RestErr {
			return nil
		}

		user, err := userService.SignUp(dto)

		assert.Nil(t, err)
		assert.EqualValues(t, dto.Email, user.Email)
	})

	t.Run("Sign_Up_Should_Throw_If_Repo_Throws", func(t *testing.T) {
		HashFunc = func(password string, cost int) ([]byte, error) {
			return []byte(""), nil
		}

		CreateFunc = func(user *User) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		user, err := userService.SignUp(dto)
		assert.EqualValues(t, "something went wrong", err.Message)
		assert.Nil(t, user)
	})

	t.Run("Sign_Up_Should_Throw_If_Can't_Hash_User", func(t *testing.T) {
		HashFunc = func(password string, cost int) ([]byte, error) {
			return nil, errors.New("")
		}

		user, err := userService.SignUp(dto)
		assert.Nil(t, user)
		assert.EqualValues(t, "something went wrong.", err.Message)
	})
}

func TestService_SignIn(t *testing.T) {
	dto := new(SignInRequestDto)
	dto.Email = "test@test.com"
	dto.Password = "123456"
	dto.RememberMe = true

	t.Run("SignIn_Should_Pass", func(t *testing.T) {
		VerifyHashFunc = func(hashedPassword, password string) error {
			return nil
		}
		CreateTokenFunc = func(userId string, rememberMe bool, authorized bool) (*token.Token, *restErrors.RestErr) {
			token := new(token.Token)
			return token, nil
		}
		GetByEmailFunc = func(email string) (*User, *restErrors.RestErr) {
			user := new(User)
			user.Email = dto.Email
			user.IsEmailVerified = true
			user.TwoFactorEnabled = false
			return user, nil
		}
		session, err := userService.SignIn(dto)
		assert.Nil(t, err)
		assert.NotNil(t, session.Token)
		assert.EqualValues(t, true, session.Authorized)
	})

	t.Run("SignIn_Authorized_Session_Should_Be_False_If_TFA_Enabled", func(t *testing.T) {
		VerifyHashFunc = func(hashedPassword, password string) error {
			return nil
		}
		GetByEmailFunc = func(email string) (*User, *restErrors.RestErr) {
			user := new(User)
			user.Email = dto.Email
			user.IsEmailVerified = true
			user.TwoFactorEnabled = true
			return user, nil
		}
		session, err := userService.SignIn(dto)
		assert.Nil(t, err)
		assert.EqualValues(t, false, session.Authorized)
	})

	t.Run("SingIn_Should_Throw_If_Email_Not_Verified", func(t *testing.T) {
		GetByEmailFunc = func(email string) (*User, *restErrors.RestErr) {
			user := new(User)
			user.Email = dto.Email
			user.IsEmailVerified = false
			return user, nil
		}

		VerifyHashFunc = func(hashedPassword, password string) error {
			return nil
		}

		session, err := userService.SignIn(dto)

		assert.Nil(t, session)
		assert.EqualValues(t, "email not verified", err.Message)
		assert.EqualValues(t, http.StatusForbidden, err.Status)
	})

	t.Run("SingIn_Should_Throw_If_User_Does't_Exit", func(t *testing.T) {
		GetByEmailFunc = func(email string) (*User, *restErrors.RestErr) {
			return nil, restErrors.NewUnAuthorizedError("invalid credentials")
		}

		session, err := userService.SignIn(dto)

		assert.Nil(t, session)
		assert.EqualValues(t, "Invalid Credentials", err.Message)
		assert.EqualValues(t, http.StatusUnauthorized, err.Status)
	})

	t.Run("SignIn_Should_Throw_If_Password_Does't_Match", func(t *testing.T) {
		GetByEmailFunc = func(email string) (*User, *restErrors.RestErr) {
			user := new(User)
			user.Email = dto.Email
			user.IsEmailVerified = true
			user.TwoFactorEnabled = false
			return user, nil
		}
		VerifyHashFunc = func(hashedPassword, password string) error {
			return errors.New("Invalid Credentials")
		}
		session, err := userService.SignIn(dto)
		assert.Nil(t, session)
		assert.EqualValues(t, "Invalid Credentials", err.Message)
		assert.EqualValues(t, http.StatusUnauthorized, err.Status)
	})

	t.Run("SignIn_Should_Throw_If_Can't_Create_Token", func(t *testing.T) {
		VerifyHashFunc = func(hashedPassword, password string) error {
			return nil
		}
		CreateTokenFunc = func(userId string, rememberMe bool, authorized bool) (*token.Token, *restErrors.RestErr) {
			return nil, restErrors.NewInternalServerError("can't create token")
		}
		GetByEmailFunc = func(email string) (*User, *restErrors.RestErr) {
			user := new(User)
			user.Email = dto.Email
			user.IsEmailVerified = true
			user.TwoFactorEnabled = false
			return user, nil
		}
		session, err := userService.SignIn(dto)
		assert.Nil(t, session)
		assert.EqualValues(t, "can't create token", err.Message)
	})

}

func TestService_GetByEmail(t *testing.T) {
	email := "test@test.com"

	t.Run("Get_By_Email_Should_Pass", func(t *testing.T) {
		GetByEmailFunc = func(email string) (*User, *restErrors.RestErr) {
			user := new(User)
			user.Email = email
			return user, nil
		}

		user, err := userService.GetByEmail(email)

		assert.Nil(t, err)
		assert.EqualValues(t, email, user.Email)
	})

	t.Run("Get_By_Email_Should_Throw_If_Does't_Exit", func(t *testing.T) {
		GetByEmailFunc = func(email string) (*User, *restErrors.RestErr) {
			return nil, restErrors.NewNotFoundError("not found")
		}

		user, err := userService.GetByEmail(email)

		assert.Nil(t, user)
		assert.EqualValues(t, http.StatusNotFound, err.Status)
		assert.EqualValues(t, "not found", err.Message)
	})
}

func TestService_VerifyEmail(t *testing.T) {

	t.Run("Verify_Email_Should_Pass", func(t *testing.T) {
		UpdateFunc = func(user *User) *restErrors.RestErr {
			return nil
		}
		user := new(User)
		err := userService.VerifyEmail(user)
		assert.Nil(t, err)
		assert.EqualValues(t, true, user.IsEmailVerified)
	})

	t.Run("Verify_Email_Should_Throw_If_Repo_Throws", func(t *testing.T) {
		UpdateFunc = func(user *User) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		user := new(User)
		err := userService.VerifyEmail(user)
		assert.EqualValues(t, "something went wrong", err.Message)
	})

}

func TestService_ResetPassword(t *testing.T) {
	user := new(User)
	t.Run("Reset_Password_Should_Pass", func(t *testing.T) {
		HashFunc = func(password string, cost int) ([]byte, error) {
			return []byte("hash"), nil
		}
		UpdateFunc = func(user *User) *restErrors.RestErr {
			return nil
		}

		err := userService.ResetPassword(user, "123456")
		assert.Nil(t, err)
		assert.EqualValues(t, "hash", user.Password)
	})

	t.Run("Reset_Password_Should_Throw_If_Repo_Throws", func(t *testing.T) {
		HashFunc = func(password string, cost int) ([]byte, error) {
			return []byte("hash"), nil
		}

		UpdateFunc = func(user *User) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		err := userService.ResetPassword(user, "123456")

		assert.EqualValues(t, "something went wrong", err.Message)
	})
	t.Run("Reset_Password_Should_Throw_Can't_Create_Hash", func(t *testing.T) {
		HashFunc = func(password string, cost int) ([]byte, error) {
			return nil, errors.New("")
		}

		UpdateFunc = func(user *User) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		err := userService.ResetPassword(user, "123456")

		assert.EqualValues(t, "something went wrong.", err.Message)
	})
}

func TestService_ChangePassword(t *testing.T) {
	user := new(User)
	dto := new(ChangePasswordRequestDto)
	dto.Password = "test"
	dto.PasswordConfirmation = "test"
	dto.OldPassword = "test"

	t.Run("Change_Password_Should_Pass", func(t *testing.T) {
		VerifyHashFunc = func(hashedPassword, password string) error {
			return nil
		}

		HashFunc = func(password string, cost int) ([]byte, error) {
			return []byte{}, nil
		}

		UpdateFunc = func(user *User) *restErrors.RestErr {
			return nil
		}

		resErr := userService.ChangePassword(user, dto)
		assert.Nil(t, resErr)
	})

	t.Run("Change_Password_Should_Throw_If_Verify_Password_Throws", func(t *testing.T) {
		VerifyHashFunc = func(hashedPassword, password string) error {
			return errors.New("")
		}

		restErr := userService.ChangePassword(user, dto)
		assert.EqualValues(t, "invalid old password", restErr.Message)
		assert.EqualValues(t, http.StatusUnauthorized, restErr.Status)
	})

	t.Run("Change_Password_Should_Throw_If_Hash_Throws", func(t *testing.T) {
		VerifyHashFunc = func(hashedPassword, password string) error {
			return nil
		}

		HashFunc = func(password string, cost int) ([]byte, error) {
			return nil, errors.New("")
		}

		restErr := userService.ChangePassword(user, dto)
		assert.EqualValues(t, "something went wrong.", restErr.Message)
		assert.EqualValues(t, http.StatusInternalServerError, restErr.Status)
	})

	t.Run("Change_Password_Should_Throw_If_Repo_Throws", func(t *testing.T) {
		HashFunc = func(password string, cost int) ([]byte, error) {
			return []byte{}, nil
		}

		VerifyHashFunc = func(hashedPassword, password string) error {
			return nil
		}

		UpdateFunc = func(user *User) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		restErr := userService.ChangePassword(user, dto)
		assert.EqualValues(t, "something went wrong", restErr.Message)
	})
}

func TestService_ChangeEmail(t *testing.T) {
	user := new(User)
	dto := new(ChangeEmailRequestDto)

	t.Run("Change_Email_Should_Pass", func(t *testing.T) {
		VerifyHashFunc = func(hashedPassword, password string) error {
			return nil
		}

		UpdateFunc = func(user *User) *restErrors.RestErr {
			return nil
		}

		err := userService.ChangeEmail(user, dto)
		assert.Nil(t, err)
	})

	t.Run("Change_Email_Should_Throw_If_Password_Is_Invalid", func(t *testing.T) {
		VerifyHashFunc = func(hashedPassword, password string) error {
			return errors.New("invalid password")
		}

		err := userService.ChangeEmail(user, dto)
		assert.EqualValues(t, "invalid password", err.Message)
		assert.EqualValues(t, http.StatusBadRequest, err.Status)
	})

	t.Run("Change_Email_Should_Throws_If_Repo_Throws", func(t *testing.T) {
		VerifyHashFunc = func(hashedPassword, password string) error {
			return nil
		}

		UpdateFunc = func(user *User) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}
		err := userService.ChangeEmail(user, dto)
		assert.EqualValues(t, "something went wrong", err.Message)
	})
}

func TestService_CreateTOTP(t *testing.T) {
	dto := new(CreateTOTPRequestDto)
	dto.Password = "123456"
	t.Run("Create_TOTP_Should_Pass", func(t *testing.T) {
		VerifyHashFunc = func(hashedPassword, password string) error {
			return nil
		}

		CreateQRCodeFunc = func(accountName string) (bytes.Buffer, string, error) {
			return bytes.Buffer{}, "", nil
		}

		UpdateFunc = func(user *User) *restErrors.RestErr {
			return nil
		}

		EncryptFunc = func(data []byte, passphrase string) (string, error) {
			return "", nil
		}

		user := new(User)
		user.Email = "test@test.com"
		buffer, err := userService.CreateTOTP(user, dto)
		assert.Nil(t, err)
		assert.NotNil(t, buffer)
		assert.NotNil(t, user.TwoFactorCipher)
	})

	t.Run("Create_TOTP_Should_Throw_If_Password_Invalid", func(t *testing.T) {
		VerifyHashFunc = func(hashedPassword, password string) error {
			return errors.New("invalid password")
		}
		user := new(User)
		_, err := userService.CreateTOTP(user, dto)
		assert.EqualValues(t, "Invalid password", err.Message)
		assert.EqualValues(t, http.StatusBadRequest, err.Status)

	})

	t.Run("Create_TOTP_Should_Throw_If_Create_Qr_Code_Throws", func(t *testing.T) {
		VerifyHashFunc = func(hashedPassword, password string) error {
			return nil
		}

		CreateQRCodeFunc = func(accountName string) (bytes.Buffer, string, error) {
			return bytes.Buffer{}, "", errors.New("")
		}
		user := new(User)
		_, err := userService.CreateTOTP(user, dto)
		assert.EqualValues(t, "something went wrong", err.Message)
	})

	t.Run("Create_TOTP_Should_Throw_If_Encryption_service_Throws", func(t *testing.T) {
		VerifyHashFunc = func(hashedPassword, password string) error {
			return nil
		}

		CreateQRCodeFunc = func(accountName string) (bytes.Buffer, string, error) {
			return bytes.Buffer{}, "", nil
		}

		EncryptFunc = func(data []byte, passphrase string) (string, error) {
			return "", errors.New("")
		}
		user := new(User)
		_, err := userService.CreateTOTP(user, dto)
		assert.EqualValues(t, "something went wrong", err.Message)
	})

	t.Run("Create_TOTP_Should_Throw_If_Repo_Throws", func(t *testing.T) {
		VerifyHashFunc = func(hashedPassword, password string) error {
			return nil
		}

		CreateQRCodeFunc = func(accountName string) (bytes.Buffer, string, error) {
			return bytes.Buffer{}, "", nil
		}

		EncryptFunc = func(data []byte, passphrase string) (string, error) {
			return "", nil
		}

		UpdateFunc = func(user *User) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		user := new(User)
		user.Email = "test@test.com"
		buffer, err := userService.CreateTOTP(user, dto)

		assert.EqualValues(t, bytes.Buffer{}, buffer)
		assert.EqualValues(t, "something went wrong", err.Message)
	})
}

func TestService_EnableTwoFactorAuth(t *testing.T) {
	t.Run("Enable_Two_Factor_Auth_Should_Pass", func(t *testing.T) {
		DecryptFunc = func(encodedCipher string, passphrase string) (string, error) {
			return "", nil
		}
		CheckOtpFunc = func(userTOTPSecret string, otp string) bool {
			return true
		}
		UpdateFunc = func(user *User) *restErrors.RestErr {
			return nil
		}
		user := new(User)
		user.TwoFactorCipher = "test"
		result, err := userService.EnableTwoFactorAuth(user, "123")

		assert.Nil(t, err)
		assert.EqualValues(t, true, result.TwoFactorEnabled)
	})

	t.Run("Enable_Two_Factor_Auth_Should_Throw_There_Is_No_Cipher_created_before", func(t *testing.T) {
		user := new(User)
		result, err := userService.EnableTwoFactorAuth(user, "123")

		assert.Nil(t, result)
		assert.EqualValues(t, "please create and register qr code first", err.Message)
	})

	t.Run("Enable_Two_Factor_Auth_Should_Throw_If_can't_Decrypt", func(t *testing.T) {
		DecryptFunc = func(encodedCipher string, passphrase string) (string, error) {
			return "", errors.New("")
		}
		user := new(User)
		user.TwoFactorCipher = "test"
		result, err := userService.EnableTwoFactorAuth(user, "123")

		assert.Nil(t, result)
		assert.EqualValues(t, "something went wrong", err.Message)
	})

	t.Run("Enable_Two_Factor_Auth_Should_Throw_If_Check_Otp_Return_False", func(t *testing.T) {
		DecryptFunc = func(encodedCipher string, passphrase string) (string, error) {
			return "", nil
		}
		CheckOtpFunc = func(userTOTPSecret string, otp string) bool {
			return false
		}

		user := new(User)
		user.TwoFactorCipher = "test"
		result, err := userService.EnableTwoFactorAuth(user, "123")
		assert.Nil(t, result)
		assert.EqualValues(t, "invalid totp code", err.Message)
	})

	t.Run("Enable_Two_Factor_Auth_Should_Throw_if_Repo_throws", func(t *testing.T) {
		DecryptFunc = func(encodedCipher string, passphrase string) (string, error) {
			return "", nil
		}
		CheckOtpFunc = func(userTOTPSecret string, otp string) bool {
			return true
		}
		UpdateFunc = func(user *User) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}
		user := new(User)
		user.TwoFactorCipher = "test"
		result, err := userService.EnableTwoFactorAuth(user, "123")

		assert.Nil(t, result)
		assert.EqualValues(t, "something went wrong", err.Message)
	})

}

func TestService_VerifyTOTP(t *testing.T) {

	t.Run("Verify_TOTP_Should_Pass", func(t *testing.T) {
		DecryptFunc = func(encodedCipher string, passphrase string) (string, error) {
			return "", nil
		}
		CheckOtpFunc = func(userTOTPSecret string, otp string) bool {
			return true
		}
		CreateTokenFunc = func(userId string, rememberMe bool, authorized bool) (*token.Token, *restErrors.RestErr) {
			return new(token.Token), nil
		}

		user := new(User)
		user.TwoFactorEnabled = true
		session, err := userService.VerifyTOTP(user, "123")

		assert.Nil(t, err)
		assert.NotNil(t, true, session.Authorized)
	})

	t.Run("Verify_TOTP_Should_Throw_If_Tfa_Disabled", func(t *testing.T) {
		user := new(User)
		user.TwoFactorEnabled = false
		session, err := userService.VerifyTOTP(user, "123")

		assert.Nil(t, session)
		assert.NotNil(t, "please enable your 2fa first", err.Message)
	})

	t.Run("Verify_TOTP_Should_Throw_If_Decrypt_throws", func(t *testing.T) {
		DecryptFunc = func(encodedCipher string, passphrase string) (string, error) {
			return "", errors.New("")
		}
		user := new(User)
		user.TwoFactorEnabled = true
		session, err := userService.VerifyTOTP(user, "123")

		assert.Nil(t, session)
		assert.NotNil(t, "something went wrong", err.Message)

	})

	t.Run("Verify_TOTP_Should_throw_If_Check_otp_returns_False", func(t *testing.T) {
		DecryptFunc = func(encodedCipher string, passphrase string) (string, error) {
			return "", nil
		}
		CheckOtpFunc = func(userTOTPSecret string, otp string) bool {
			return false
		}

		user := new(User)
		user.TwoFactorEnabled = true
		session, err := userService.VerifyTOTP(user, "123")

		assert.Nil(t, session)
		assert.NotNil(t, "invalid totp code", err.Message)
	})

	t.Run("Verify_TOTP_Should_Throw_If_Create_Token_Throws", func(t *testing.T) {
		DecryptFunc = func(encodedCipher string, passphrase string) (string, error) {
			return "", nil
		}
		CheckOtpFunc = func(userTOTPSecret string, otp string) bool {
			return true
		}
		CreateTokenFunc = func(userId string, rememberMe bool, authorized bool) (*token.Token, *restErrors.RestErr) {
			return nil, restErrors.NewInternalServerError("can't create token")
		}

		user := new(User)
		user.TwoFactorEnabled = true
		session, err := userService.VerifyTOTP(user, "123")

		assert.Nil(t, session)
		assert.EqualValues(t, "can't create token", err.Message)
	})

}

func TestService_DisableTwoFactorAuth(t *testing.T) {
	dto := new(DisableTOTPRequestDto)
	dto.Password = "123456"
	t.Run("Disable_Tfa_Should_Pass", func(t *testing.T) {
		VerifyHashFunc = func(hashedPassword, password string) error {
			return nil
		}

		UpdateFunc = func(user *User) *restErrors.RestErr {
			return nil
		}

		user := new(User)
		user.TwoFactorEnabled = true
		err := userService.DisableTwoFactorAuth(user, dto)

		assert.Nil(t, err)
		assert.EqualValues(t, false, user.TwoFactorEnabled)
	})

	t.Run("Disable_Tfa_Should_Throw_If_Password_isInvalid", func(t *testing.T) {
		VerifyHashFunc = func(hashedPassword, password string) error {
			return errors.New("invalid password")
		}
		user := new(User)
		user.TwoFactorEnabled = true
		err := userService.DisableTwoFactorAuth(user, dto)

		assert.EqualValues(t, "Invalid password", err.Message)
		assert.EqualValues(t, http.StatusBadRequest, err.Status)
	})
	t.Run("Disable_Tfa_Should_Throw_If_Already_Disabled", func(t *testing.T) {
		VerifyHashFunc = func(hashedPassword, password string) error {
			return nil
		}

		UpdateFunc = func(user *User) *restErrors.RestErr {
			return nil
		}

		user := new(User)
		user.TwoFactorEnabled = false
		err := userService.DisableTwoFactorAuth(user, dto)

		assert.EqualValues(t, "2fa already disabled", err.Message)
	})

	t.Run("Disable_Tfa_Should_Throw_If_Repo_Throws", func(t *testing.T) {
		VerifyHashFunc = func(hashedPassword, password string) error {
			return nil
		}

		UpdateFunc = func(user *User) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		user := new(User)
		user.TwoFactorEnabled = true
		err := userService.DisableTwoFactorAuth(user, dto)

		assert.EqualValues(t, "something went wrong", err.Message)
	})
}

func TestService_FindWhereIdInSlice(t *testing.T) {
	t.Run("find_users_where_id_in_slice_should_pass", func(t *testing.T) {
		FindWhereIdInSliceFunc = func(ids []string) ([]*User, *restErrors.RestErr) {
			return []*User{new(User), new(User)}, nil
		}
		list, err := userService.FindWhereIdInSlice([]string{})
		assert.Nil(t, err)
		assert.Len(t, list, 2)
	})
}

func TestService_Count(t *testing.T) {
	t.Run("count should pass", func(t *testing.T) {
		CountFunc = func() (int64, *restErrors.RestErr) {
			return 1, nil
		}
		count, err := userService.Count()
		assert.EqualValues(t, 1, count)
		assert.Nil(t, err)
	})

	t.Run("count should throw if repo throws", func(t *testing.T) {
		CountFunc = func() (int64, *restErrors.RestErr) {
			return 0, restErrors.NewInternalServerError("something went wrong")
		}
		count, err := userService.Count()
		assert.EqualValues(t, 0, count)
		assert.EqualValues(t, "something went wrong", err.Message)
	})
}
func TestService_SetAsPlatformAdmin(t *testing.T) {
	user := new(User)

	t.Run("Set as platform admin should pass", func(t *testing.T) {
		UpdateFunc = func(user *User) *restErrors.RestErr {
			return nil
		}

		resErr := userService.SetAsPlatformAdmin(user)
		assert.Nil(t, resErr)
		assert.EqualValues(t, true, user.PlatformAdmin)
	})

	t.Run("Change_Password_Should_Throw_If_Repo_Throws", func(t *testing.T) {
		UpdateFunc = func(user *User) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		restErr := userService.SetAsPlatformAdmin(user)
		assert.EqualValues(t, "something went wrong", restErr.Message)
	})
}
