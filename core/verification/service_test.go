package verification

import (
	"errors"
	"github.com/kotalco/core-api/config"
	"os"
	"testing"
	"time"

	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	verificationService IService
	CreateFunc          func(verification *Verification) restErrors.IRestErr
	GetByUserIdFunc     func(userId string) (*Verification, restErrors.IRestErr)
	UpdateFunc          func(verification *Verification) restErrors.IRestErr
	WithTransactionFunc func(txHandle *gorm.DB) IRepository

	HashFunc       func(password string, cost int) ([]byte, error)
	VerifyHashFunc func(hashedPassword, password string) error
)

type verificationRepositoryMock struct{}

func (vRepository verificationRepositoryMock) WithoutTransaction() IRepository {
	return vRepository
}

type hashingServiceMock struct{}

// verification repository methods
func (vRepository verificationRepositoryMock) WithTransaction(txHandle *gorm.DB) IRepository {
	return vRepository
}

func (verificationRepositoryMock) Create(verification *Verification) restErrors.IRestErr {
	return CreateFunc(verification)
}
func (verificationRepositoryMock) GetByUserId(userId string) (*Verification, restErrors.IRestErr) {
	return GetByUserIdFunc(userId)
}
func (verificationRepositoryMock) Update(verification *Verification) restErrors.IRestErr {
	return UpdateFunc(verification)
}

// hashing service methods
func (hashingServiceMock) Hash(password string, cost int) ([]byte, error) {
	return HashFunc(password, cost)
}

func (hashingServiceMock) VerifyHash(hashedPassword, password string) error {
	return VerifyHashFunc(hashedPassword, password)
}

func TestMain(m *testing.M) {
	verificationRepository = &verificationRepositoryMock{}
	hashing = &hashingServiceMock{}
	verificationService = NewService()
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
			config.Environment.VerificationTokenLength = "invalid"
			token, err := generateToken()
			assert.EqualValues(t, "", token)
			assert.EqualValues(t, "some thing went wrong", err.Error())
			config.Environment.VerificationTokenLength = "80"
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
			assert.EqualValues(t, err.Error(), "some thing went wrong")
		})
	})
}

func TestService_Create(t *testing.T) {
	t.Run("Create_Should_Pass", func(t *testing.T) {
		generateToken = func() (string, restErrors.IRestErr) {
			return "token", nil
		}

		hashToken = func(token string) (string, restErrors.IRestErr) {
			return "hash", nil
		}

		CreateFunc = func(verification *Verification) restErrors.IRestErr {
			return nil
		}

		result, err := verificationService.Create("test")
		assert.Nil(t, err)
		assert.EqualValues(t, "token", result)
	})

	t.Run("Create_Should_Throw_if_Generate_Token_Throws", func(t *testing.T) {
		generateToken = func() (string, restErrors.IRestErr) {
			return "", restErrors.NewInternalServerError("something went wrong")
		}

		result, err := verificationService.Create("1")

		assert.EqualValues(t, "", result)
		assert.EqualValues(t, "something went wrong", err.Error())
	})

	t.Run("Create_Should_Throw_if_Hash_Token_Throws", func(t *testing.T) {
		generateToken = func() (string, restErrors.IRestErr) {
			return "", nil
		}

		hashToken = func(token string) (string, restErrors.IRestErr) {
			return "", restErrors.NewInternalServerError("something went wrong")
		}

		result, err := verificationService.Create("1")

		assert.EqualValues(t, "", result)
		assert.EqualValues(t, "something went wrong", err.Error())
	})

	t.Run("Create_Should_Throw_if_Repository_Throws", func(t *testing.T) {
		generateToken = func() (string, restErrors.IRestErr) {
			return "", nil
		}

		hashToken = func(token string) (string, restErrors.IRestErr) {
			return "", nil
		}

		CreateFunc = func(verification *Verification) restErrors.IRestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		result, err := verificationService.Create("1")

		assert.EqualValues(t, "", result)
		assert.EqualValues(t, "something went wrong", err.Error())
	})

	t.Run("Create_Should_Throw_If_Verification_Token_Expiry_Is_Invalid", func(t *testing.T) {
		generateToken = func() (string, restErrors.IRestErr) {
			return "token", nil
		}

		hashToken = func(token string) (string, restErrors.IRestErr) {
			return "hash", nil
		}
		current := config.Environment.VerificationTokenExpiryHours
		config.Environment.VerificationTokenExpiryHours = "invalid"
		result, err := verificationService.Create("test")
		assert.EqualValues(t, "", result)
		assert.EqualValues(t, "something went wrong", err.Error())
		config.Environment.VerificationTokenExpiryHours = current
	})

}

func TestService_GetByUserId(t *testing.T) {
	t.Run("Get_Verification_By_User_Id", func(t *testing.T) {
		GetByUserIdFunc = func(userId string) (*Verification, restErrors.IRestErr) {
			return new(Verification), nil
		}

		verification, err := verificationService.GetByUserId("123")

		assert.Nil(t, err)
		assert.NotNil(t, verification)
	})

	t.Run("Get_Verification_By_User_Id_Should_Throw_If_Repo_Throws", func(t *testing.T) {
		GetByUserIdFunc = func(userId string) (*Verification, restErrors.IRestErr) {
			return nil, restErrors.NewInternalServerError("")
		}

		verification, err := verificationService.GetByUserId("123")

		assert.Nil(t, verification)
		assert.EqualValues(t, "No such verification", err.Error())

	})
}

func TestService_Verify(t *testing.T) {
	t.Run("Verify_Should_Pass", func(t *testing.T) {
		verification := new(Verification)
		verification.ExpiresAt = time.Now().UTC().Add(time.Duration(1) * time.Hour).Unix()

		GetByUserIdFunc = func(userId string) (*Verification, restErrors.IRestErr) {
			return verification, nil
		}

		VerifyHashFunc = func(hashedPassword, password string) error {
			return nil
		}

		UpdateFunc = func(verification *Verification) restErrors.IRestErr {
			return nil
		}

		err := verificationService.Verify("1", "hash")

		assert.Nil(t, err)
		assert.EqualValues(t, true, verification.Completed)

	})
	t.Run("Verify_Should_Throw_If_Get_By_User_Id_Throws", func(t *testing.T) {
		GetByUserIdFunc = func(userId string) (*Verification, restErrors.IRestErr) {
			return nil, restErrors.NewInternalServerError("")
		}

		err := verificationService.Verify("1", "hash")
		assert.EqualValues(t, "No such verification", err.Error())
	})
	t.Run("Verify_Should_Throw_If_token_Expired", func(t *testing.T) {
		verification := new(Verification)
		verification.ExpiresAt = 0

		GetByUserIdFunc = func(userId string) (*Verification, restErrors.IRestErr) {
			return verification, nil
		}

		err := verificationService.Verify("1", "hash")

		assert.EqualValues(t, "token expired", err.Error())
	})
	t.Run("Verify_Should_Throw_If_User_Already_Verified", func(t *testing.T) {
		verification := new(Verification)
		verification.ExpiresAt = time.Now().UTC().Add(time.Duration(1) * time.Hour).Unix()

		verification.Completed = true
		GetByUserIdFunc = func(userId string) (*Verification, restErrors.IRestErr) {
			return verification, nil
		}

		err := verificationService.Verify("1", "hash")

		assert.EqualValues(t, "user already verified", err.Error())
	})
	t.Run("Verify_Should_Throw_If_Token_Is_Invalid", func(t *testing.T) {
		verification := new(Verification)
		verification.ExpiresAt = time.Now().UTC().Add(time.Duration(1) * time.Hour).Unix()

		GetByUserIdFunc = func(userId string) (*Verification, restErrors.IRestErr) {
			return verification, nil
		}

		VerifyHashFunc = func(hashedPassword, password string) error {
			return errors.New("")
		}

		err := verificationService.Verify("1", "hash")

		assert.EqualValues(t, "invalid token", err.Error())
	})
	t.Run("Verify_Should_Throw_If_Repo_Throws", func(t *testing.T) {
		verification := new(Verification)
		verification.ExpiresAt = time.Now().UTC().Add(time.Duration(1) * time.Hour).Unix()

		GetByUserIdFunc = func(userId string) (*Verification, restErrors.IRestErr) {
			return verification, nil
		}

		VerifyHashFunc = func(hashedPassword, password string) error {
			return nil
		}

		UpdateFunc = func(verification *Verification) restErrors.IRestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		err := verificationService.Verify("1", "hash")

		assert.EqualValues(t, "something went wrong", err.Error())
	})
}

func TestService_Resend(t *testing.T) {
	verification := new(Verification)

	t.Run("Resend_Should_Pass", func(t *testing.T) {

		GetByUserIdFunc = func(userId string) (*Verification, restErrors.IRestErr) {
			return verification, nil
		}

		generateToken = func() (string, restErrors.IRestErr) {
			return "token", nil
		}

		hashToken = func(token string) (string, restErrors.IRestErr) {
			return "hash", nil
		}
		expiryDate = func() (int64, restErrors.IRestErr) {
			return time.Now().UTC().Add(time.Duration(1) * time.Hour).Unix(), nil
		}

		UpdateFunc = func(verification *Verification) restErrors.IRestErr {
			return nil
		}

		token, err := verificationService.Resend("123")

		assert.Nil(t, err)
		assert.EqualValues(t, "token", token)
	})

	t.Run("Resend_Should_Pass_And_Create_New_Verification_If_It_The_First_Time", func(t *testing.T) {

		GetByUserIdFunc = func(userId string) (*Verification, restErrors.IRestErr) {
			return nil, restErrors.NewNotFoundError("no such verification")
		}
		generateToken = func() (string, restErrors.IRestErr) {
			return "token", nil
		}

		hashToken = func(token string) (string, restErrors.IRestErr) {
			return "hash", nil
		}
		expiryDate = func() (int64, restErrors.IRestErr) {
			return time.Now().UTC().Add(time.Duration(1) * time.Hour).Unix(), nil
		}

		CreateFunc = func(verification *Verification) restErrors.IRestErr {
			return nil
		}

		token, err := verificationService.Resend("123")

		assert.Nil(t, err)
		assert.EqualValues(t, "token", token)
	})

	t.Run("Resend_Should_Throw_If_Get_User_By_Id", func(t *testing.T) {
		GetByUserIdFunc = func(userId string) (*Verification, restErrors.IRestErr) {
			return nil, restErrors.NewInternalServerError("something went wrong")
		}

		token, err := verificationService.Resend("1")

		assert.EqualValues(t, "", token)
		assert.EqualValues(t, "something went wrong", err.Error())
	})

	t.Run("Resend_Should_Throw_If_Can't_Generate_Token", func(t *testing.T) {
		GetByUserIdFunc = func(userId string) (*Verification, restErrors.IRestErr) {
			return verification, nil
		}

		generateToken = func() (string, restErrors.IRestErr) {
			return "", restErrors.NewInternalServerError("something went wrong")
		}

		token, err := verificationService.Resend("123")

		assert.EqualValues(t, "", token)
		assert.EqualValues(t, "something went wrong", err.Error())
	})

	t.Run("Resend_Should_Throw_If_Can't_Hash_Token", func(t *testing.T) {

		GetByUserIdFunc = func(userId string) (*Verification, restErrors.IRestErr) {
			return verification, nil
		}

		generateToken = func() (string, restErrors.IRestErr) {
			return "token", nil
		}

		hashToken = func(token string) (string, restErrors.IRestErr) {
			return "", restErrors.NewInternalServerError("can't hash token")
		}

		token, err := verificationService.Resend("123")

		assert.EqualValues(t, "", token)
		assert.EqualValues(t, "can't hash token", err.Error())
	})

	t.Run("Resend_Should_Throw_If_Can't_create_expiry_date", func(t *testing.T) {

		GetByUserIdFunc = func(userId string) (*Verification, restErrors.IRestErr) {
			return verification, nil
		}

		generateToken = func() (string, restErrors.IRestErr) {
			return "token", nil
		}

		hashToken = func(token string) (string, restErrors.IRestErr) {
			return "", nil
		}
		expiryDate = func() (int64, restErrors.IRestErr) {
			return 0, restErrors.NewInternalServerError("something went wrong")
		}

		token, err := verificationService.Resend("123")

		assert.EqualValues(t, "", token)
		assert.EqualValues(t, "something went wrong", err.Error())
	})

	t.Run("Resend_Should_Throw_If_Repo_Throws", func(t *testing.T) {

		GetByUserIdFunc = func(userId string) (*Verification, restErrors.IRestErr) {
			return verification, nil
		}

		generateToken = func() (string, restErrors.IRestErr) {
			return "token", nil
		}

		hashToken = func(token string) (string, restErrors.IRestErr) {
			return "hash", nil
		}
		expiryDate = func() (int64, restErrors.IRestErr) {
			return time.Now().UTC().Add(time.Duration(1) * time.Hour).Unix(), nil
		}

		UpdateFunc = func(verification *Verification) restErrors.IRestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		token, err := verificationService.Resend("")

		assert.EqualValues(t, "", token)
		assert.EqualValues(t, "something went wrong", err.Error())
	})

}
