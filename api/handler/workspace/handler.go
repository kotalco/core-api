package workspace

import (
	"github.com/gofiber/fiber/v2"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/api/pkg/shared"
	"github.com/kotalco/cloud-api/internal/workspace"
	"github.com/kotalco/cloud-api/pkg/k8s"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	"github.com/kotalco/cloud-api/pkg/token"
	"net/http"
)

var (
	workspaceService = workspace.NewService()
	namespaceService = k8s.NewNamespaceService()
)

//Create validate dto , create new workspace, creates new namespace in k8
func Create(c *fiber.Ctx) error {

	userId := c.Locals("user").(token.UserDetails).ID

	dto := new(workspace.CreateWorkspaceRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.Status).JSON(badReq)
	}

	err := workspace.Validate(dto)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	txHandle := sqlclient.Begin()
	model, err := workspaceService.WithTransaction(txHandle).Create(dto, userId)
	if err != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(err.Status).JSON(err)
	}

	err = namespaceService.Create(model.K8sNamespace)
	if err != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(err.Status).JSON(err)
	}

	sqlclient.Commit(txHandle)

	return c.Status(http.StatusCreated).JSON(shared.NewResponse(new(workspace.WorkspaceResponseDto).Marshall(model)))

}

//Update validate dto , validate user authenticity & update workspace
func Update(c *fiber.Ctx) error {

	model := c.Locals("workspace").(workspace.Workspace)

	dto := new(workspace.UpdateWorkspaceRequestDto)
	dto.ID = model.ID

	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.Status).JSON(badReq)
	}

	err := workspace.Validate(dto)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	txHandle := sqlclient.Begin()
	err = workspaceService.WithTransaction(txHandle).Update(dto, &model)
	if err != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(err.Status).JSON(err)
	}
	sqlclient.Commit(txHandle)

	return c.Status(http.StatusOK).JSON(shared.NewResponse(new(workspace.WorkspaceResponseDto).Marshall(&model)))
}

//Delete deletes user workspace and associated namespace
func Delete(c *fiber.Ctx) error {
	userId := c.Locals("user").(token.UserDetails).ID
	workspaceId := c.Params("id")

	model, err := workspaceService.GetById(workspaceId)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	if model.UserId != userId {
		err := restErrors.NewNotFoundError("record not found")
		return c.Status(err.Status).JSON(err)
	}

	txHandle := sqlclient.Begin()
	err = workspaceService.WithTransaction(txHandle).Delete(model)
	if err != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(err.Status).JSON(err)
	}

	err = namespaceService.Delete(model.K8sNamespace)
	if err != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(err.Status).JSON(err)
	}

	sqlclient.Commit(txHandle)

	return c.SendStatus(http.StatusNoContent)
}
