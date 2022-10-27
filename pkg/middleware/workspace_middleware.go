package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/internal/workspace"
	"github.com/kotalco/cloud-api/pkg/token"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"net/http"
)

var workspaceRepo = workspace.NewRepository()

// IsWorkspace check if user exits in the workspace, creates workspace, workspaceUser locals
// used to protect cloud-api handlers that's need workspace
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

	return c.Next()
}

// DeploymentsWorkspaceProtected validate if user exist in the workspace, creates namespace local to be used by the community-api
// used to protect deployments handlers
func DeploymentsWorkspaceProtected(c *fiber.Ctx) error {
	userId := c.Locals("user").(token.UserDetails).ID
	var model *workspace.Workspace
	var workspaceId string
	var err *restErrors.RestErr

	//get workspaceId
	if c.Request().Header.IsPost() { //if method is post verb,  workspace expected to be a body field ,or it's going to be  default
		var bodyFields map[string]interface{}
		_ = c.BodyParser(&bodyFields)
		if bodyFields["workspace_id"] != nil {
			workspaceId = bodyFields["workspace_id"].(string)
		}
	} else { // if method IS NOT post verb, workspace expected to be qs , or it's going to be default
		if c.Query("workspace_id") != "" {
			workspaceId = c.Query("workspace_id")
		}
	}

	//get workspace model
	if workspaceId != "" {
		model, err = workspaceRepo.GetById(workspaceId)
		if err != nil {
			return c.Status(err.Status).JSON(err)
		}
	} else {
		model, err = workspaceRepo.GetByNamespace("default")
		if err != nil {
			return c.Status(err.Status).JSON(err)
		}
	}

	//check if user exist in the workspace
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

	c.Locals("namespace", model.K8sNamespace)
	return c.Next()
}
