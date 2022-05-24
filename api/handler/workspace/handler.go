package workspace

import (
	"github.com/gofiber/fiber/v2"
	restErrors "github.com/kotalco/api/pkg/errors"
	"github.com/kotalco/api/pkg/shared"
	"github.com/kotalco/cloud-api/internal/user"
	"github.com/kotalco/cloud-api/internal/workspace"
	"github.com/kotalco/cloud-api/pkg/k8s"
	"github.com/kotalco/cloud-api/pkg/sendgrid"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	"github.com/kotalco/cloud-api/pkg/token"
	"net/http"
)

var (
	workspaceService = workspace.NewService()
	namespaceService = k8s.NewNamespaceService()
	userService      = user.NewService()
	mailService      = sendgrid.NewService()
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
	model := c.Locals("workspace").(workspace.Workspace)

	txHandle := sqlclient.Begin()
	err := workspaceService.WithTransaction(txHandle).Delete(&model)
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

//GetByUserId find workspaces by userId
func GetByUserId(c *fiber.Ctx) error {
	userId := c.Locals("user").(token.UserDetails).ID

	list, err := workspaceService.GetByUserId(userId)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	var marshalled = make([]workspace.WorkspaceResponseDto, 0)
	for _, v := range list {
		record := new(workspace.WorkspaceResponseDto).Marshall(v)
		marshalled = append(marshalled, *record)
	}

	return c.Status(http.StatusOK).JSON(shared.NewResponse(marshalled))
}

//AddMember adds new member to workspace
func AddMember(c *fiber.Ctx) error {
	dto := new(workspace.AddWorkspaceMemberDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.Status).JSON(badReq)
	}

	err := workspace.Validate(dto)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	member, err := userService.GetByEmail(dto.Email)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	model := c.Locals("workspace").(workspace.Workspace)

	txHandle := sqlclient.Begin()
	err = workspaceService.WithTransaction(txHandle).AddWorkspaceMember(member.ID, &model)
	if err != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(err.Status).JSON(err)
	}

	sqlclient.Commit(txHandle)

	mailRequestDto := new(sendgrid.WorkspaceInvitationMailRequestDto)
	mailRequestDto.Email = dto.Email
	mailRequestDto.WorkspaceName = model.Name
	mailRequestDto.WorkspaceId = model.ID
	go mailService.WorkspaceInvitation(mailRequestDto)

	return c.Status(http.StatusOK).JSON(shared.NewResponse(shared.SuccessMessage{
		Message: "user has been added to the workspace",
	}))
}
