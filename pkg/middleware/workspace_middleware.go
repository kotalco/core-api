package middleware

import (
	"github.com/gofiber/fiber/v2"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/cloud-api/internal/workspace"
	"github.com/kotalco/cloud-api/pkg/token"
	"net/http"
)

var workspaceRepo = workspace.NewRepository()

func IsWorkspace(c *fiber.Ctx) error {
	workspaceId := c.Params("id")
	userId := c.Locals("user").(token.UserDetails).ID

	model, err := workspaceRepo.GetById(workspaceId)
	if err != nil {
		if err.Status == http.StatusNotFound {
			notFoundErr := restErrors.NewNotFoundError("no such record")
			return c.Status(notFoundErr.Status).JSON(notFoundErr)
		}
		return c.Status(err.Status).JSON(err)
	}

	validUser := false
	for _, v := range model.WorkspaceUsers {
		if v.UserId == userId {
			validUser = true
			c.Locals("workspaceUser", v)
			break
		}
	}
	if !validUser {
		notFoundErr := restErrors.NewNotFoundError("no such record")
		return c.Status(notFoundErr.Status).JSON(notFoundErr)
	}

	c.Locals("workspace", *model)

	c.Next()
	return nil
}
