package verification

import (
	"fmt"
	"github.com/kotalco/core-api/config"
	"net/http"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/logger"
	"github.com/kotalco/core-api/pkg/security"
)

type service struct{}

type IService interface {
	WithTransaction(txHandle *gorm.DB) IService
	WithoutTransaction() IService
	Create(userId string) (string, restErrors.IRestErr)
	GetByUserId(userId string) (*Verification, restErrors.IRestErr)
	Resend(userId string) (string, restErrors.IRestErr)
	Verify(userId string, token string) restErrors.IRestErr
}

var (
	verificationRepository = NewRepository()
	hashing                = security.NewHashing()
)

func NewService() IService {
	newService := &service{}
	return newService
}

func (vService service) WithTransaction(txHandle *gorm.DB) IService {
	verificationRepository = verificationRepository.WithTransaction(txHandle)
	return vService
}
func (vService service) WithoutTransaction() IService {
	verificationRepository = verificationRepository.WithoutTransaction()
	return vService
}

// Create creates a new verification token for the user
func (service) Create(userId string) (string, restErrors.IRestErr) {
	token, restErr := generateToken()
	if restErr != nil {
		return "", restErr
	}

	hashedToken, restErr := hashToken(token)
	if restErr != nil {
		return "", restErr
	}

	verification := new(Verification)
	verification.ID = uuid.New().String()
	verification.Token = hashedToken
	verification.UserId = userId
	expiresAt, err := expiryDate()
	if err != nil {
		return "", err
	}
	verification.ExpiresAt = expiresAt

	restErr = verificationRepository.Create(verification)
	if restErr != nil {
		return "", restErr
	}

	return token, nil
}

// GetByUserId get verification by user id
func (service) GetByUserId(userId string) (*Verification, restErrors.IRestErr) {
	verification, err := verificationRepository.GetByUserId(userId)
	if err != nil {
		return nil, restErrors.NewNotFoundError(fmt.Sprintf("No such verification"))
	}

	return verification, err
}

// Verify verify token hash
func (service) Verify(userId string, token string) restErrors.IRestErr {
	verification, err := verificationRepository.GetByUserId(userId)
	if err != nil {
		return restErrors.NewNotFoundError(fmt.Sprintf("No such verification"))
	}

	if verification.ExpiresAt < time.Now().Unix() {
		return restErrors.NewBadRequestError("token expired")
	}

	if verification.Completed {
		return restErrors.NewBadRequestError("user already verified")
	}

	verifyErr := hashing.VerifyHash(verification.Token, token)

	if verifyErr != nil {
		return restErrors.NewBadRequestError("invalid token")
	}

	verification.Completed = true

	err = verificationRepository.Update(verification)
	if err != nil {
		return err
	}

	return nil
}

// Resend  verification token to user
func (vService service) Resend(userId string) (string, restErrors.IRestErr) {
	verification, err := verificationRepository.GetByUserId(userId)
	if err != nil {
		if err.StatusCode() == http.StatusNotFound {
			return vService.Create(userId)
		}
		return "", err
	}

	token, err := generateToken()
	if err != nil {
		return "", err
	}
	hashedToken, err := hashToken(token)
	if err != nil {
		return "", err
	}

	verification.Completed = false
	verification.Token = hashedToken
	expiresAt, err := expiryDate()
	if err != nil {
		return "", err
	}
	verification.ExpiresAt = expiresAt
	err = verificationRepository.Update(verification)
	if err != nil {
		return "", err
	}

	return token, nil
}

// generateToken creates a random token to the user which will be sent to user email
var generateToken = func() (string, restErrors.IRestErr) {
	tokenLength, err := strconv.Atoi(config.Environment.VerificationTokenLength)
	if err != nil {
		go logger.Error("generateToken", err)
		return "", restErrors.NewInternalServerError("some thing went wrong")
	}

	token := security.GenerateRandomString(tokenLength)

	return token, nil
}

// hashToken hashes the verification token before sending it to the user email
var hashToken = func(token string) (string, restErrors.IRestErr) {

	hashedToken, err := hashing.Hash(token, 6)
	if err != nil {
		go logger.Error("hashToken", err)
		return "", restErrors.NewInternalServerError("some thing went wrong")
	}

	stringifyToken := string(hashedToken)

	return stringifyToken, nil
}

var expiryDate = func() (int64, restErrors.IRestErr) {
	tokenExpires, convErr := strconv.Atoi(config.Environment.VerificationTokenExpiryHours)
	if convErr != nil {
		go logger.Warn("VERIFICATION_EXPIRY_DATE", convErr)
		return 0, restErrors.NewInternalServerError("something went wrong")
	}
	return time.Now().UTC().Add(time.Duration(tokenExpires) * time.Hour).Unix(), nil
}
