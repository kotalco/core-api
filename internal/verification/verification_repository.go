package verification

import (
	"fmt"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	"gorm.io/gorm"
	"regexp"
)

type verificationRepository struct{}

type verificationRepositoryInterface interface {
	Create(verification *Verification) *restErrors.RestErr
	GetByUserId(userId string) (*Verification, *restErrors.RestErr)
	Update(verification *Verification) *restErrors.RestErr
}

var (
	VerificationRepository verificationRepositoryInterface
	dbClient               *gorm.DB
)

func init() {
	VerificationRepository = &verificationRepository{}
	dbClient, _ = sqlclient.OpenDBConnection()
}

func (repo verificationRepository) Create(verification *Verification) *restErrors.RestErr {
	res := dbClient.Create(verification)
	if res.Error != nil {
		duplicateEmail, _ := regexp.Match("duplicate key", []byte(res.Error.Error()))
		if duplicateEmail {
			return restErrors.NewBadRequestError("email already exits")
		}
		go logger.Error(repo.Create, res.Error)
		return restErrors.NewInternalServerError("can't create verification")
	}

	return nil
}

func (repo verificationRepository) GetByUserId(userId string) (*Verification, *restErrors.RestErr) {
	var verification = new(Verification)

	result := dbClient.Where("user_id = ?", userId).First(verification)
	if result.Error != nil {
		return nil, restErrors.NewNotFoundError(fmt.Sprintf("can't find verification with userId  %s", userId))
	}

	return verification, nil
}

func (repo verificationRepository) Update(verification *Verification) *restErrors.RestErr {
	resp := dbClient.Save(verification)
	if resp.Error != nil {
		go logger.Error(repo.Update, resp.Error)
	}

	return nil
}
