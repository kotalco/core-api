package middleware

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/internal/workspace"
	"github.com/kotalco/cloud-api/pkg/token"
	restErrors "github.com/kotalco/community-api/pkg/errors"
)

func IsNamespace(c *fiber.Ctx) error {
	var workspaceRepo = workspace.NewRepository()
	var namespace = c.Query("namespace") //namespace exits as query string

	userId := c.Locals("user").(token.UserDetails).ID

	//namespace exits as a body field
	if namespace == "" { //namespace exits as body filed
		body := make(map[string]interface{})
		json.Unmarshal(c.Body(), &body)
		_, ok := body["namespace"]
		if ok { //user supported the namespace
			namespace = body["namespace"].(string)
		} else { // user didn't support the namespace
			namespace = "default"
		}
	}

	model, err := workspaceRepo.GetByNamespace(namespace)
	if err != nil {
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
