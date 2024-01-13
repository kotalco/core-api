package token

import (
	"fmt"
	"github.com/kotalco/cloud-api/config"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	restErrors "github.com/kotalco/cloud-api/pkg/errors"
	"github.com/kotalco/cloud-api/pkg/logger"
)

type token struct{}

type IToken interface {
	CreateToken(userId string, rememberMe bool, authorized bool) (*Token, restErrors.IRestErr)
	ExtractTokenMetadata(bearToken string) (*AccessDetails, restErrors.IRestErr)
}

func NewToken() IToken {
	newToken := &token{}
	return newToken
}

func (token) CreateToken(userId string, rememberMe bool, authorized bool) (*Token, restErrors.IRestErr) {
	var tokenExpires int
	var convErr error
	tokenExpires, convErr = strconv.Atoi(config.Environment.JwtSecretKeyExpireHoursCount)
	if rememberMe {
		tokenExpires, convErr = strconv.Atoi(config.Environment.JwtSecretKeyExpireHoursCountRememberMe)
	}

	if convErr != nil {
		go logger.Error("INVALID_TOKEN_EXPIRY", convErr)
		return nil, restErrors.NewInternalServerError("some thing went wrong")
	}
	t := new(Token)
	t.Expires = time.Now().UTC().Add(time.Duration(tokenExpires) * time.Hour).Unix()
	t.TokenUuid = uuid.New().String()
	t.Authorized = authorized
	//Creating Access Token
	tClaims := jwt.MapClaims{}
	tClaims["authorized"] = t.Authorized
	tClaims["access_uuid"] = t.TokenUuid
	tClaims["user_id"] = userId
	tClaims["exp"] = t.Expires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, tClaims)
	var err error
	t.AccessToken, err = at.SignedString([]byte(config.Environment.AccessSecret))
	if err != nil {
		go logger.Error("CREATE_TOKEN_GENERATOR", err)
		return nil, restErrors.NewInternalServerError("some thing went wrong")
	}
	return t, nil
}

func (token) ExtractTokenMetadata(bearToken string) (*AccessDetails, restErrors.IRestErr) {
	token, err := verifyToken(bearToken)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUuid, ok := claims["access_uuid"].(string)
		if !ok {
			return nil, err
		}
		return &AccessDetails{
			TokenUuid:  accessUuid,
			UserId:     claims["user_id"].(string),
			Authorized: claims["authorized"].(bool),
		}, nil
	}
	return nil, err
}

func verifyToken(bearToken string) (*jwt.Token, restErrors.IRestErr) {
	tokenString := extractToken(bearToken)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token_generator method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Environment.AccessSecret), nil
	})
	if err != nil {
		go logger.Error("VERIFY_TOKEN", err)
		return nil, restErrors.NewUnAuthorizedError("invalid token")
	}
	return token, nil
}

// ExtractToken  from the request body
func extractToken(bearToken string) string {
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}
