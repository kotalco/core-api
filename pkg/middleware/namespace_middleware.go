package middleware

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/cloud-api/internal/workspace"
	"github.com/kotalco/cloud-api/pkg/token"
)

func IsNamespace(c *fiber.Ctx) error {
	var workspaceRepo = workspace.NewRepository()
	var namespace = c.Query("namespace") //namespace exits as query string

	userId := c.Locals("user").(token.UserDetails).ID

	//namespace exits as a body field
	if namespace == "" { //namespace exits as body filed
		body := make(map[string]interface{})
		err := json.Unmarshal(c.Body(), &body)
		if err != nil {
			logger.Error("IS_NAMESPACE", err)
			internalErr := restErrors.NewInternalServerError("something went wrong")
			return c.Status(internalErr.Status).JSON(internalErr)
		}

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
