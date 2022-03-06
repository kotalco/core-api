package middleware

import (
	"github.com/gofiber/fiber/v2"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/cloud-api/internal/user"
	"github.com/kotalco/cloud-api/pkg/tokens"
	"net/http"
)

var AuthorizedUser user.User

func JWTProtected(c *fiber.Ctx) error {
	BearerToken := c.Get("Authorization")
	accessDetails, err := tokens.ExtractTokenMetadata(BearerToken)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	user, err := user.UserRepository.GetById(accessDetails.UserId)
	if err != nil {
		if err.Status == http.StatusNotFound {
			unAuthErr := restErrors.NewUnAuthorizedError("no such user")
			return c.Status(unAuthErr.Status).JSON(unAuthErr)
		}
		return c.Status(err.Status).JSON(err)
	}
	AuthorizedUser = *user
	c.Next()
	return nil
}
