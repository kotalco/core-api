package verification

import (
	"errors"
	"fmt"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/cloud-api/pkg/config"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"os"
	"testing"
	"time"
)

var (
	verificationService IService
	CreateFunc          func(verification *Verification) *restErrors.RestErr
	GetByUserIdFunc     func(userId string) (*Verification, *restErrors.RestErr)
	UpdateFunc          func(verification *Verification) *restErrors.RestErr
	WithTransactionFunc func(txHandle *gorm.DB) IRepository

	HashFunc       func(password string, cost int) ([]byte, error)
	VerifyHashFunc func(hashedPassword, password string) error
)

type verificationRepositoryMock struct{}
type hashingServiceMock struct{}

//verification repository methods
func (vRepository verificationRepositoryMock) WithTransaction(txHandle *gorm.DB) IRepository {
	return vRepository
}

func (verificationRepositoryMock) Create(verification *Verification) *restErrors.RestErr {
	return CreateFunc(verification)
}
func (verificationRepositoryMock) GetByUserId(userId string) (*Verification, *restErrors.RestErr) {
	return GetByUserIdFunc(userId)
}
func (verificationRepositoryMock) Update(verification *Verification) *restErrors.RestErr {
	return UpdateFunc(verification)
}

//hashing service methods
func (hashingServiceMock) Hash(password string, cost int) ([]byte, error) {
	return HashFunc(password, cost)
}

func (hashingServiceMock) VerifyHash(hashedPassword, password string) error {
	return VerifyHashFunc(hashedPassword, password)
}

func TestMain(m *testing.M) {
	verificationService = NewService()
	verificationRepository = &verificationRepositoryMock{}
	hashing = &hashingServiceMock{}
	code := m.Run()
	os.Exit(code)
}

func TestService_Helper_Function_Before_They_Get_Mocked(t *testing.T) {
	t.Run("Generate_Token_Helper_Func", func(t *testing.T) {
		t.Run("Generate_Token_Should_Pass", func(t *testing.T) {
			token, err := generateToken()
			assert.Nil(t, err)
			assert.Len(t, token, 80)
		})

		t.Run("Generate_Token_Should_Throw_If_Verification_Token_Conf_Is_Invalid", func(t *testing.T) {
			config.EnvironmentConf["VERIFICATION_TOKEN_LENGTH"] = "invalid"
			token, err := generateToken()
			assert.EqualValues(t, "", token)
			assert.EqualValues(t, "some thing went wrong", err.Message)
			config.EnvironmentConf["VERIFICATION_TOKEN_LENGTH"] = "80"
		})
	})

	t.Run("Hash_Token_Helper_Func", func(t *testing.T) {
		t.Run("Hash_Token_Should_Pass", func(t *testing.T) {
			HashFunc = func(password string, cost int) ([]byte, error) {
				return []byte(""), nil
			}

			hash, err := hashToken("token")

			assert.Nil(t, err)
			assert.NotNil(t, hash)
		})

		t.Run("Hash_Token_Should_Throw_If_Hashing_Service_Throws", func(t *testing.T) {
			HashFunc = func(password string, cost int) ([]byte, error) {
				return nil, errors.New("")
			}

			hash, err := hashToken("token")

			assert.EqualValues(t, "", hash)
			assert.EqualValues(t, err.Message, "some thing went wrong")
		})
	})
}

func TestService_Create(t *testing.T) {
	t.Run("Create_Should_Pass", func(t *testing.T) {
		generateToken = func() (string, *restErrors.RestErr) {
			return "token", nil
		}

		hashToken = func(token string) (string, *restErrors.RestErr) {
			return "hash", nil
		}

		CreateFunc = func(verification *Verification) *restErrors.RestErr {
			return nil
		}

		result, err := verificationService.Create("test")
		assert.Nil(t, err)
		assert.EqualValues(t, "token", result)
	})

	t.Run("Create_Should_Throw_if_Generate_Token_Throws", func(t *testing.T) {
		generateToken = func() (string, *restErrors.RestErr) {
			return "", restErrors.NewInternalServerError("something went wrong")
		}

		result, err := verificationService.Create("1")

		assert.EqualValues(t, "", result)
		assert.EqualValues(t, "something went wrong", err.Message)
	})

	t.Run("Create_Should_Throw_if_Hash_Token_Throws", func(t *testing.T) {
		generateToken = func() (string, *restErrors.RestErr) {
			return "", nil
		}

		hashToken = func(token string) (string, *restErrors.RestErr) {
			return "", restErrors.NewInternalServerError("something went wrong")
		}

		result, err := verificationService.Create("1")

		assert.EqualValues(t, "", result)
		assert.EqualValues(t, "something went wrong", err.Message)
	})

	t.Run("Create_Should_Throw_if_Repository_Throws", func(t *testing.T) {
		generateToken = func() (string, *restErrors.RestErr) {
			return "", nil
		}

		hashToken = func(token string) (string, *restErrors.RestErr) {
			return "", nil
		}

		CreateFunc = func(verification *Verification) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		result, err := verificationService.Create("1")

		assert.EqualValues(t, "", result)
		assert.EqualValues(t, "something went wrong", err.Message)
	})

	t.Run("Create_Should_Throw_If_Verification_Token_Expiry_Is_Invalid", func(t *testing.T) {
		generateToken = func() (string, *restErrors.RestErr) {
			return "token", nil
		}

		hashToken = func(token string) (string, *restErrors.RestErr) {
			return "hash", nil
		}
		current := config.EnvironmentConf["VERIFICATION_TOKEN_EXPIRY_HOURS"]
		config.EnvironmentConf["VERIFICATION_TOKEN_EXPIRY_HOURS"] = "invalid"
		result, err := verificationService.Create("test")
		assert.EqualValues(t, "", result)
		assert.EqualValues(t, "something went wrong", err.Message)
		config.EnvironmentConf["VERIFICATION_TOKEN_EXPIRY_HOURS"] = current
		fmt.Println(current)
	})

}

func TestService_GetByUserId(t *testing.T) {
	t.Run("Get_Verification_By_User_Id", func(t *testing.T) {
		GetByUserIdFunc = func(userId string) (*Verification, *restErrors.RestErr) {
			return new(Verification), nil
		}

		verification, err := verificationService.GetByUserId("123")

		assert.Nil(t, err)
		assert.NotNil(t, verification)
	})

	t.Run("Get_Verification_By_User_Id_Should_Throw_If_Repo_Throws", func(t *testing.T) {
		GetByUserIdFunc = func(userId string) (*Verification, *restErrors.RestErr) {
			return nil, restErrors.NewInternalServerError("")
		}

		verification, err := verificationService.GetByUserId("123")

		assert.Nil(t, verification)
		assert.EqualValues(t, "No such verification", err.Message)

	})
}

func TestService_Verify(t *testing.T) {
	t.Run("Verify_Should_Pass", func(t *testing.T) {
		verification := new(Verification)
		verification.ExpiresAt = time.Now().UTC().Add(time.Duration(1) * time.Hour).Unix()

		GetByUserIdFunc = func(userId string) (*Verification, *restErrors.RestErr) {
			return verification, nil
		}

		VerifyHashFunc = func(hashedPassword, password string) error {
			return nil
		}

		UpdateFunc = func(verification *Verification) *restErrors.RestErr {
			return nil
		}

		err := verificationService.Verify("1", "hash")

		assert.Nil(t, err)
		assert.EqualValues(t, true, verification.Completed)

	})
	t.Run("Verify_Should_Throw_If_Get_By_User_Id_Throws", func(t *testing.T) {
		GetByUserIdFunc = func(userId string) (*Verification, *restErrors.RestErr) {
			return nil, restErrors.NewInternalServerError("")
		}

		err := verificationService.Verify("1", "hash")
		assert.EqualValues(t, "No such verification", err.Message)
	})
	t.Run("Verify_Should_Throw_If_token_Expired", func(t *testing.T) {
		verification := new(Verification)
		verification.ExpiresAt = 0

		GetByUserIdFunc = func(userId string) (*Verification, *restErrors.RestErr) {
			return verification, nil
		}

		err := verificationService.Verify("1", "hash")

		assert.EqualValues(t, "token expired", err.Message)
	})
	t.Run("Verify_Should_Throw_If_User_Already_Verified", func(t *testing.T) {
		verification := new(Verification)
		verification.ExpiresAt = time.Now().UTC().Add(time.Duration(1) * time.Hour).Unix()

		verification.Completed = true
		GetByUserIdFunc = func(userId string) (*Verification, *restErrors.RestErr) {
			return verification, nil
		}

		err := verificationService.Verify("1", "hash")

		assert.EqualValues(t, "user already verified", err.Message)
	})
	t.Run("Verify_Should_Throw_If_Token_Is_Invalid", func(t *testing.T) {
		verification := new(Verification)
		verification.ExpiresAt = time.Now().UTC().Add(time.Duration(1) * time.Hour).Unix()

		GetByUserIdFunc = func(userId string) (*Verification, *restErrors.RestErr) {
			return verification, nil
		}

		VerifyHashFunc = func(hashedPassword, password string) error {
			return errors.New("")
		}

		err := verificationService.Verify("1", "hash")

		assert.EqualValues(t, "invalid token", err.Message)
	})
	t.Run("Verify_Should_Throw_If_Repo_Throws", func(t *testing.T) {
		verification := new(Verification)
		verification.ExpiresAt = time.Now().UTC().Add(time.Duration(1) * time.Hour).Unix()

		GetByUserIdFunc = func(userId string) (*Verification, *restErrors.RestErr) {
			return verification, nil
		}

		VerifyHashFunc = func(hashedPassword, password string) error {
			return nil
		}

		UpdateFunc = func(verification *Verification) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		err := verificationService.Verify("1", "hash")

		assert.EqualValues(t, "something went wrong", err.Message)
	})
}

func TestService_Resend(t *testing.T) {
	verification := new(Verification)
	verification.ExpiresAt = time.Now().UTC().Add(time.Duration(1) * time.Hour).Unix()

	t.Run("Resend_Should_Pass", func(t *testing.T) {

		GetByUserIdFunc = func(userId string) (*Verification, *restErrors.RestErr) {
			return verification, nil
		}

		generateToken = func() (string, *restErrors.RestErr) {
			return "token", nil
		}

		hashToken = func(token string) (string, *restErrors.RestErr) {
			return "hash", nil
		}

		UpdateFunc = func(verification *Verification) *restErrors.RestErr {
			return nil
		}

		token, err := verificationService.Resend("123")

		assert.Nil(t, err)
		assert.EqualValues(t, "token", token)
	})

	t.Run("Resend_Should_Pass_And_Create_New_Verification_If_It_The_First_Time", func(t *testing.T) {

		GetByUserIdFunc = func(userId string) (*Verification, *restErrors.RestErr) {
			return nil, restErrors.NewNotFoundError("no such verification")
		}
		generateToken = func() (string, *restErrors.RestErr) {
			return "token", nil
		}

		hashToken = func(token string) (string, *restErrors.RestErr) {
			return "hash", nil
		}

		CreateFunc = func(verification *Verification) *restErrors.RestErr {
			return nil
		}

		token, err := verificationService.Resend("123")

		assert.Nil(t, err)
		assert.EqualValues(t, "token", token)
	})

	t.Run("Resend_Should_Throw_If_Get_User_By_Id", func(t *testing.T) {
		GetByUserIdFunc = func(userId string) (*Verification, *restErrors.RestErr) {
			return nil, restErrors.NewInternalServerError("something went wrong")
		}

		token, err := verificationService.Resend("1")

		assert.EqualValues(t, "", token)
		assert.EqualValues(t, "something went wrong", err.Message)
	})

	t.Run("Resend_Should_Throw_If_Can't_Generate_Token", func(t *testing.T) {
		GetByUserIdFunc = func(userId string) (*Verification, *restErrors.RestErr) {
			return verification, nil
		}

		generateToken = func() (string, *restErrors.RestErr) {
			return "", restErrors.NewInternalServerError("something went wrong")
		}

		token, err := verificationService.Resend("123")

		assert.EqualValues(t, "", token)
		assert.EqualValues(t, "something went wrong", err.Message)
	})

	t.Run("Resend_Should_Throw_If_Can't_Hash_Token", func(t *testing.T) {

		GetByUserIdFunc = func(userId string) (*Verification, *restErrors.RestErr) {
			return verification, nil
		}

		generateToken = func() (string, *restErrors.RestErr) {
			return "token", nil
		}

		hashToken = func(token string) (string, *restErrors.RestErr) {
			return "", restErrors.NewInternalServerError("can't hash token")
		}

		token, err := verificationService.Resend("123")

		assert.EqualValues(t, "", token)
		assert.EqualValues(t, "can't hash token", err.Message)
	})

	t.Run("Resend_Should_Throw_If_Repo_Throws", func(t *testing.T) {

		GetByUserIdFunc = func(userId string) (*Verification, *restErrors.RestErr) {
			return verification, nil
		}

		generateToken = func() (string, *restErrors.RestErr) {
			return "token", nil
		}

		hashToken = func(token string) (string, *restErrors.RestErr) {
			return "hash", nil
		}

		UpdateFunc = func(verification *Verification) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		token, err := verificationService.Resend("")

		assert.EqualValues(t, "", token)
		assert.EqualValues(t, "something went wrong", err.Message)
	})

}
