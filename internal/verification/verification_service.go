package verification

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/cloud-api/pkg/config"
	"github.com/kotalco/cloud-api/pkg/security"
)

type verificationService struct{}

type verificationServiceInterface interface {
	Create(userId string) (string, *restErrors.RestErr)
	GetByUserId(userId string) (*Verification, *restErrors.RestErr)
	Resend(userId string) (string, *restErrors.RestErr)
	Verify(userId string, token string) *restErrors.RestErr
}

var (
	VerificationService verificationServiceInterface
)

func init() { VerificationService = &verificationService{} }

//Create creates a new verification token for the user
func (service verificationService) Create(userId string) (string, *restErrors.RestErr) {
	token, restErr := generateToken()
	if restErr != nil {
		return "", restErr
	}

	hashedToken, restErr := hashToken(token)
	if restErr != nil {
		return "", restErr
	}

	tokenExpires, convErr := strconv.Atoi(config.EnvironmentConf["VERIFICATION_TOKEN_EXPIRY_HOURS"])
	if convErr != nil {
		go logger.Error(service.Create, convErr)
		return "", restErrors.NewInternalServerError("something went wrong")
	}

	verification := new(Verification)
	verification.ID = uuid.New().String()
	verification.Token = *hashedToken
	verification.UserId = userId
	verification.ExpiresAt = time.Now().UTC().Add(time.Duration(tokenExpires) * time.Hour).Unix()

	restErr = VerificationRepository.Create(verification)
	if restErr != nil {
		return "", restErr
	}

	return token, nil
}

//GetByUserId get verification by user id
func (service verificationService) GetByUserId(userId string) (*Verification, *restErrors.RestErr) {
	verification, err := VerificationRepository.GetByUserId(userId)
	if err != nil {
		return nil, restErrors.NewNotFoundError(fmt.Sprintf("No such verification"))
	}

	return verification, err
}

//Verify verify token hash
func (service verificationService) Verify(userId string, token string) *restErrors.RestErr {
	verification, err := VerificationRepository.GetByUserId(userId)
	if err != nil {
		return restErrors.NewNotFoundError(fmt.Sprintf("No such verification"))
	}

	if verification.Completed {
		return restErrors.NewBadRequestError("user already verified")
	}

	if verification.ExpiresAt < time.Now().Unix() {
		return restErrors.NewBadRequestError("token expired")
	}

	verifyErr := security.VerifyPassword(verification.Token, token)

	if verifyErr != nil {
		return restErrors.NewBadRequestError("invalid token")
	}

	verification.Completed = true

	err = VerificationRepository.Update(verification)
	if err != nil {
		return err
	}

	return nil
}

//Resend  verification token to user
func (service verificationService) Resend(userId string) (string, *restErrors.RestErr) {
	verification, err := VerificationRepository.GetByUserId(userId)
	if err != nil {
		if err.Status == http.StatusNotFound {
			return service.Create(userId)
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
	verification.Token = *hashedToken
	err = VerificationRepository.Update(verification)
	if err != nil {
		return "", err
	}

	return token, nil
}

//generateToken creates a random token to the user which will be send to user email
func generateToken() (string, *restErrors.RestErr) {
	tokenLength, err := strconv.Atoi(config.EnvironmentConf["VERIFICATION_TOKEN_LENGTH"])
	if err != nil {
		go logger.Error(generateToken, err)
		return "", restErrors.NewInternalServerError("some thing went wrong")
	}

	token := security.GenerateRandomString(tokenLength)

	return token, nil
}

//hashToken hashes the verification token before sending it to the user email
func hashToken(token string) (*string, *restErrors.RestErr) {
	hashedToken, err := security.Hash(token, 6)
	if err != nil {
		go logger.Error(generateToken, err)
		return nil, restErrors.NewInternalServerError("some thing went wrong")
	}

	stringifyToken := string(hashedToken)

	return &stringifyToken, nil
}
