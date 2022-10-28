package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/internal/workspace"
	"github.com/kotalco/cloud-api/pkg/token"
	restErrors "github.com/kotalco/community-api/pkg/errors"
)

var workspaceRepo = workspace.NewRepository()

// IsWorkspace checks for workspace_id passed as query string to get the workspace model, if the id isn't passed it gets the default workspace, creates workspace local
func IsWorkspace(c *fiber.Ctx) error {
	var model *workspace.Workspace
	var err *restErrors.RestErr

	//get workspace model
	if c.Query("workspace_id") != "" {
		model, err = workspaceRepo.GetById(c.Query("workspace_id"))
		if err != nil {
			return c.Status(err.Status).JSON(err)
		}
	} else {
		model, err = workspaceRepo.GetByNamespace("default")
		if err != nil {
			return c.Status(err.Status).JSON(err)
		}
	}

	c.Locals("workspace", *model)
	return c.Next()
}

// DeploymentsWorkspaceProtected checks for the workspace_id passed as query string or as a body field to get the workspace model, if not passed it gets the default workspace
// creates workspace model local, and namespace name local
func DeploymentsWorkspaceProtected(c *fiber.Ctx) error {
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

	c.Locals("workspace", *model)
	c.Locals("namespace", model.K8sNamespace)
	return c.Next()
}

// ValidateWorkspaceMembership validates if the user belongs to the current workspace, creates workspaceUser local
func ValidateWorkspaceMembership(c *fiber.Ctx) error {
	workspaceModel := c.Locals("workspace").(workspace.Workspace)
	userId := c.Locals("user").(token.UserDetails).ID

	//check if user exist in the workspace
	validUser := false
	for _, v := range workspaceModel.WorkspaceUsers {
		if v.UserId == userId {
			validUser = true
			c.Locals("workspaceUser", v)
			break
		}
	}
	if !validUser {
		notFoundErr := restErrors.NewForbiddenError("you ain't a member of this workspace ")
		return c.Status(notFoundErr.Status).JSON(notFoundErr)
	}

	return c.Next()
}
