package verification

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/kotalco/cloud-api/pkg/config"
	"github.com/kotalco/cloud-api/pkg/security"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
)

type service struct{}

type IService interface {
	WithTransaction(txHandle *gorm.DB) IService
	Create(userId string) (string, *restErrors.RestErr)
	GetByUserId(userId string) (*Verification, *restErrors.RestErr)
	Resend(userId string) (string, *restErrors.RestErr)
	Verify(userId string, token string) *restErrors.RestErr
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

// Create creates a new verification token for the user
func (service) Create(userId string) (string, *restErrors.RestErr) {
	token, restErr := generateToken()
	if restErr != nil {
		return "", restErr
	}

	hashedToken, restErr := hashToken(token)
	if restErr != nil {
		return "", restErr
	}

	tokenExpires, convErr := strconv.Atoi(config.Environment.VerificationTokenExpiryHours)
	if convErr != nil {
		go logger.Error(service.Create, convErr)
		return "", restErrors.NewInternalServerError("something went wrong")
	}

	verification := new(Verification)
	verification.ID = uuid.New().String()
	verification.Token = hashedToken
	verification.UserId = userId
	verification.ExpiresAt = time.Now().UTC().Add(time.Duration(tokenExpires) * time.Hour).Unix()

	restErr = verificationRepository.Create(verification)
	if restErr != nil {
		return "", restErr
	}

	return token, nil
}

// GetByUserId get verification by user id
func (service) GetByUserId(userId string) (*Verification, *restErrors.RestErr) {
	verification, err := verificationRepository.GetByUserId(userId)
	if err != nil {
		return nil, restErrors.NewNotFoundError(fmt.Sprintf("No such verification"))
	}

	return verification, err
}

// Verify verify token hash
func (service) Verify(userId string, token string) *restErrors.RestErr {
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
func (vService service) Resend(userId string) (string, *restErrors.RestErr) {
	verification, err := verificationRepository.GetByUserId(userId)
	if err != nil {
		if err.Status == http.StatusNotFound {
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
	err = verificationRepository.Update(verification)
	if err != nil {
		return "", err
	}

	return token, nil
}

// generateToken creates a random token to the user which will be sent to user email
var generateToken = func() (string, *restErrors.RestErr) {
	tokenLength, err := strconv.Atoi(config.Environment.VerificationTokenLength)
	if err != nil {
		go logger.Error("generateToken", err)
		return "", restErrors.NewInternalServerError("some thing went wrong")
	}

	token := security.GenerateRandomString(tokenLength)

	return token, nil
}

// hashToken hashes the verification token before sending it to the user email
var hashToken = func(token string) (string, *restErrors.RestErr) {

	hashedToken, err := hashing.Hash(token, 6)
	if err != nil {
		go logger.Error("hashToken", err)
		return "", restErrors.NewInternalServerError("some thing went wrong")
	}

	stringifyToken := string(hashedToken)

	return stringifyToken, nil
}
