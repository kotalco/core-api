package middleware

import (
	"github.com/gofiber/fiber/v2"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/cloud-api/internal/workspaceuser"
	"github.com/kotalco/cloud-api/pkg/roles"
	"github.com/kotalco/cloud-api/pkg/token"
	"k8s.io/apimachinery/pkg/util/json"
)

func IsAdmin(c *fiber.Ctx) error {

	workspaceUser, err := validateWorkspaceUser(c)
	if err != nil {
		logger.Panic("IS_ADMIN_MIDDLEWARE", err)
		internalErr := restErrors.NewInternalServerError("something went wrong")
		return c.Status(internalErr.Status).JSON(internalErr)
	}
	if workspaceUser.Role != roles.Admin {
		forbidden := restErrors.NewForbiddenError("unAuthorized action")
		return c.Status(forbidden.Status).JSON(forbidden)
	}
	c.Next()
	return nil
}

func IsWriter(c *fiber.Ctx) error {
	workspaceUser, err := validateWorkspaceUser(c)
	if err != nil {
		logger.Panic("IS_ADMIN_MIDDLEWARE", err)
		internalErr := restErrors.NewInternalServerError("something went wrong")
		return c.Status(internalErr.Status).JSON(internalErr)
	}
	if workspaceUser.Role != roles.Admin && workspaceUser.Role != roles.Writer {
		forbidden := restErrors.NewForbiddenError("unAuthorized action")
		return c.Status(forbidden.Status).JSON(forbidden)
	}
	c.Next()
	return nil
}

func IsReader(c *fiber.Ctx) error {
	workspaceUser, err := validateWorkspaceUser(c)
	if err != nil {
		logger.Panic("IS_ADMIN_MIDDLEWARE", err)
		internalErr := restErrors.NewInternalServerError("something went wrong")
		return c.Status(internalErr.Status).JSON(internalErr)
	}

	if workspaceUser.Role != roles.Admin && workspaceUser.Role != roles.Writer && workspaceUser.Role != roles.Reader {
		forbidden := restErrors.NewForbiddenError("unAuthorized action")
		return c.Status(forbidden.Status).JSON(forbidden)
	}
	c.Next()
	return nil
}

func validateWorkspaceUser(c *fiber.Ctx) (workspaceuser.WorkspaceUser, *restErrors.RestErr) {
	var workspaceUser workspaceuser.WorkspaceUser
	var namespace = "default"
	userId := c.Locals("user").(token.UserDetails).ID

	if c.Locals("workspaceUser") != nil { //check for workspace user working with cloud api handlers
		workspaceUser = c.Locals("workspaceUser").(workspaceuser.WorkspaceUser)
	} else { //working with api sdk
		if c.Query("namespace") == "" { //check if namespace appended as body field (create requests)
			body := make(map[string]interface{})
			err := json.Unmarshal(c.Body(), &body)
			if err != nil {
				logger.Info("didn't pass namespace")
				namespace = c.Params("namespace", "default")
			} else {
				namespace = body["namespace"].(string)
			}
		} else {
			namespace = c.Params("namespace")
		}

	}

	model, err := workspaceRepo.GetByNamespace(namespace)
	if err != nil {
		logger.Panic("IS_ADMIN_MIDDLEWARE", err)
		internalErr := restErrors.NewInternalServerError("something went wrong")
		return workspaceuser.WorkspaceUser{}, internalErr
	}

	validUser := false
	for _, v := range model.WorkspaceUsers {
		if v.UserId == userId {
			validUser = true
			workspaceUser = v
			break
		}
	}
	if !validUser {
		return workspaceuser.WorkspaceUser{}, restErrors.NewNotFoundError("no such record")

	}
	return workspaceUser, nil
}
