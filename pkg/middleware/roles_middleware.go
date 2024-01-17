package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/core-api/core/workspaceuser"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/roles"
)

func IsAdmin(c *fiber.Ctx) error {
	workspaceUser := c.Locals("workspaceUser").(workspaceuser.WorkspaceUser)

	if workspaceUser.Role != roles.Admin {
		forbidden := restErrors.NewForbiddenError("unAuthorized action")
		return c.Status(forbidden.StatusCode()).JSON(forbidden)
	}
	c.Next()
	return nil
}

func IsWriter(c *fiber.Ctx) error {
	workspaceUser := c.Locals("workspaceUser").(workspaceuser.WorkspaceUser)

	if workspaceUser.Role != roles.Admin && workspaceUser.Role != roles.Writer {
		forbidden := restErrors.NewForbiddenError("unAuthorized action")
		return c.Status(forbidden.StatusCode()).JSON(forbidden)
	}
	c.Next()
	return nil
}

func IsReader(c *fiber.Ctx) error {
	workspaceUser := c.Locals("workspaceUser").(workspaceuser.WorkspaceUser)

	if workspaceUser.Role != roles.Admin && workspaceUser.Role != roles.Writer && workspaceUser.Role != roles.Reader {
		forbidden := restErrors.NewForbiddenError("unAuthorized action")
		return c.Status(forbidden.StatusCode()).JSON(forbidden)
	}
	c.Next()
	return nil
}
