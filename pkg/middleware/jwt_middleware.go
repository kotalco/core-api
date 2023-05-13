package middleware

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/internal/user"
	"github.com/kotalco/cloud-api/pkg/token"
	restErrors "github.com/kotalco/community-api/pkg/errors"
)

var userRepository = user.NewRepository()
var tokenService = token.NewToken()

func JWTProtected(c *fiber.Ctx) error {
	BearerToken := c.Get("Authorization", c.Query("authorization"))

	accessDetails, err := tokenService.ExtractTokenMetadata(BearerToken)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}
	user, err := userRepository.GetById(accessDetails.UserId)
	if err != nil {
		if err.StatusCode() == http.StatusNotFound {
			unAuthErr := restErrors.NewUnAuthorizedError("no such user")
			return c.Status(unAuthErr.StatusCode()).JSON(unAuthErr)
		}
		return c.Status(err.StatusCode()).JSON(err)
	}
	userDetails := new(token.UserDetails)
	userDetails.ID = user.ID
	c.Locals("user", *userDetails)

	c.Next()
	return nil
}
