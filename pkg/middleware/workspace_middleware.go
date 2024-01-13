package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/core/workspace"
	restErrors "github.com/kotalco/cloud-api/pkg/errors"
	"github.com/kotalco/cloud-api/pkg/token"
)

var workspaceRepo = workspace.NewRepository()

// WorkspaceProtected checks for the workspace_id passed as query string or as a body field to get the workspace model, if not passed it gets the default workspace
// creates workspace model local, and namespace name local
func WorkspaceProtected(c *fiber.Ctx) error {
	var model *workspace.Workspace
	var workspaceId string
	var err restErrors.IRestErr

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
			return c.Status(err.StatusCode()).JSON(err)
		}
	} else {
		model, err = workspaceRepo.GetByNamespace("default")
		if err != nil {
			return c.Status(err.StatusCode()).JSON(err)
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
		return c.Status(notFoundErr.StatusCode()).JSON(notFoundErr)
	}

	return c.Next()
}
