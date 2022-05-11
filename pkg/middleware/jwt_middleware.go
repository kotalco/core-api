package middleware

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/cloud-api/internal/user"
	"github.com/kotalco/cloud-api/pkg/token"
)

var userRepository = user.NewRepository()
var tokenService = token.NewToken()

func JWTProtected(c *fiber.Ctx) error {
	BearerToken := c.Get("Authorization")
	accessDetails, err := tokenService.ExtractTokenMetadata(BearerToken)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	user, err := userRepository.GetById(accessDetails.UserId)
	if err != nil {
		if err.Status == http.StatusNotFound {
			unAuthErr := restErrors.NewUnAuthorizedError("no such user")
			return c.Status(unAuthErr.Status).JSON(unAuthErr)
		}
		return c.Status(err.Status).JSON(err)
	}
	c.Locals("user", token.AuthorizedUser(*user))
	c.Next()
	return nil
}
