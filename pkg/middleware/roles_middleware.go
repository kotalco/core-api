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
		logger.Error("IS_ADMIN_MIDDLEWARE", err)
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
		logger.Error("IS_Writer_MIDDLEWARE", err)
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
		logger.Error("IS_READER_MIDDLEWARE", err)
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

// validateWorkspaceUser get the current workspace the user working with either by passing the workspace id as an id
// in workspace module actions which will result having a local under then name workspace and workspaceUser under the name workspaceUser
//or getting the workspace and workspaceUser for the deployments by the passed namespace as QS or passed as body field
func validateWorkspaceUser(c *fiber.Ctx) (workspaceuser.WorkspaceUser, *restErrors.RestErr) {
	var workspaceUser workspaceuser.WorkspaceUser
	var namespace = "default"
	userId := c.Locals("user").(token.UserDetails).ID

	if c.Locals("workspaceUser") != nil { //check for workspace user working with cloud api handlers
		workspaceUser = c.Locals("workspaceUser").(workspaceuser.WorkspaceUser)
		return workspaceUser, nil
	} else { //working with api sdk
		if c.Query("namespace") == "" { //check if namespace appended as body field (create Deployment requests) )
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
		logger.Error("VALIDATE_WORKSPACE_USER_IN_ROLES_MIDDLEWARE", err)
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
