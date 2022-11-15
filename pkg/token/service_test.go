package token

import (
	"testing"

	"github.com/kotalco/cloud-api/pkg/config"
	"github.com/stretchr/testify/assert"
)

var tokenTestingService = NewToken()

func TestToken_CreateToken(t *testing.T) {
	t.Run("Create_token_Should_Pass", func(t *testing.T) {
		newToken, err := tokenTestingService.CreateToken("1", false, true)
		assert.Nil(t, err)
		assert.True(t, newToken.Authorized)
		assert.NotNil(t, newToken.AccessToken)
	})

	t.Run("Create_Token_Should_Fail_If_Token_Expiry_Is_Invalid", func(t *testing.T) {
		oldConf := config.Environment.JwtSecretKeyExpireHoursCount
		config.Environment.JwtSecretKeyExpireHoursCount = "invalid"
		newToken, err := tokenTestingService.CreateToken("1", false, true)
		assert.Nil(t, newToken)
		assert.EqualValues(t, "some thing went wrong", err.Message)
		config.Environment.JwtSecretKeyExpireHoursCount = oldConf

	})
}

func TestToken_ExtractTokenMetadata(t *testing.T) {
	t.Run("Extract_Token_Meta_Data_Should_Pass", func(t *testing.T) {
		newToken, err := tokenTestingService.CreateToken("1", false, true)
		assert.Nil(t, err)
		accessDetails, err := tokenTestingService.ExtractTokenMetadata("Bearer " + newToken.AccessToken)
		assert.Nil(t, err)
		assert.EqualValues(t, "1", accessDetails.UserId)
		assert.EqualValues(t, true, accessDetails.Authorized)
	})

	t.Run("Extract_Token_Meta_Data_Should_Fail_If_Token_invalid", func(t *testing.T) {
		newToken, err := tokenTestingService.CreateToken("1", false, true)
		assert.Nil(t, err)
		accessDetails, err := tokenTestingService.ExtractTokenMetadata("Bearer " + newToken.AccessToken + "invalid")
		assert.Nil(t, accessDetails)
		assert.EqualValues(t, "invalid token", err.Message)
	})
	t.Run("Extract_Token_Meta_Data_Should_Fail_If_Token_invalid", func(t *testing.T) {
		newToken, err := tokenTestingService.CreateToken("1", false, true)
		assert.Nil(t, err)
		accessDetails, err := tokenTestingService.ExtractTokenMetadata(newToken.AccessToken + "invalid")
		assert.Nil(t, accessDetails)
		assert.EqualValues(t, "invalid token", err.Message)
	})

}
