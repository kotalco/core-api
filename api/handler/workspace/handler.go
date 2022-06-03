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

	for _, v := range model.WorkspaceUsers {
		if v.UserId == member.ID {
			conflictErr := restErrors.NewConflictError("User is already a member of the workspace")
			return c.Status(conflictErr.Status).JSON(conflictErr)
		}
	}

	txHandle := sqlclient.Begin()
	err = workspaceService.WithTransaction(txHandle).AddWorkspaceMember(&model, member.ID)
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

//Leave removes workspace member from workspace
func Leave(c *fiber.Ctx) error {
	model := c.Locals("workspace").(workspace.Workspace)
	userId := c.Locals("user").(token.UserDetails).ID

	if model.UserId == userId {
		err := restErrors.NewForbiddenError("you can't leave your own workspace")
		return c.Status(err.Status).JSON(err)
	}

	txHandle := sqlclient.Begin()
	err := workspaceService.WithTransaction(txHandle).DeleteWorkspaceMember(&model, userId)
	if err != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(err.Status).JSON(err)
	}

	sqlclient.Commit(txHandle)

	return c.Status(http.StatusOK).JSON(shared.NewResponse(shared.SuccessMessage{
		Message: "You're no longer member of this workspace",
	}))
}

//RemoveMember workspace owner removes workspace member form his/her workspace
func RemoveMember(c *fiber.Ctx) error {
	model := c.Locals("workspace").(workspace.Workspace)
	userId := c.Locals("user").(token.UserDetails).ID
	memberId := c.Params("user_id")

	if model.UserId != userId { //check if the user is the owner
		err := restErrors.NewForbiddenError("you can only delete other users from your own workspace")
		return c.Status(err.Status).JSON(err)
	}

	if model.UserId == memberId { //check if the-to-be deleted user isn't the owner of the workspace
		err := restErrors.NewForbiddenError("you can't leave your own workspace")
		return c.Status(err.Status).JSON(err)
	}

	exist := false //check if the to-be-delete user exists in the workspace
	for _, v := range model.WorkspaceUsers {
		if v.UserId == memberId {
			exist = true
			break
		}
	}
	if !exist {
		notFoundErr := restErrors.NewNotFoundError("user isn't a member of the workspace")
		return c.Status(notFoundErr.Status).JSON(notFoundErr)
	}

	txHandle := sqlclient.Begin()
	err := workspaceService.WithTransaction(txHandle).DeleteWorkspaceMember(&model, memberId)
	if err != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(err.Status).JSON(err)
	}

	sqlclient.Commit(txHandle)

	return c.Status(http.StatusOK).JSON(shared.NewResponse(shared.SuccessMessage{
		Message: "User has been removed from workspace",
	}))
}

//Members returns a list of workspace members
func Members(c *fiber.Ctx) error {
	model := c.Locals("workspace").(workspace.Workspace)
	userIds := make([]string, len(model.WorkspaceUsers))
	for k, v := range model.WorkspaceUsers {
		userIds[k] = v.UserId
	}

	workspaceMembersList, err := userService.FindWhereIdInSlice(userIds)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	result := make([]user.PublicUserResponseDto, len(workspaceMembersList))
	for k, v := range workspaceMembersList {
		result[k] = new(user.PublicUserResponseDto).Marshall(v)
	}

	return c.Status(http.StatusOK).JSON(shared.NewResponse(result))
}
