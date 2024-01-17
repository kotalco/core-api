package workspace

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kotalco/core-api/core/user"
	"github.com/kotalco/core-api/core/workspace"
	"github.com/kotalco/core-api/core/workspaceuser"
	"github.com/kotalco/core-api/k8s"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/pagination"
	"github.com/kotalco/core-api/pkg/roles"
	"github.com/kotalco/core-api/pkg/sendgrid"
	"github.com/kotalco/core-api/pkg/sqlclient"
	"github.com/kotalco/core-api/pkg/token"
	"net/http"
)

var (
	workspaceService = workspace.NewService()
	namespaceService = k8s.NewNamespaceService()
	userService      = user.NewService()
	mailService      = sendgrid.NewService()
)

// Create validate dto , create new workspace, creates new namespace in k8
func Create(c *fiber.Ctx) error {

	userId := c.Locals("user").(token.UserDetails).ID

	dto := new(workspace.CreateWorkspaceRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}

	err := workspace.Validate(dto)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	txHandle := sqlclient.Begin()
	model, err := workspaceService.WithTransaction(txHandle).Create(dto, userId, uuid.NewString())
	if err != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(err.StatusCode()).JSON(err)
	}

	err = namespaceService.Create(model.K8sNamespace)
	if err != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(err.StatusCode()).JSON(err)
	}

	sqlclient.Commit(txHandle)

	return c.Status(http.StatusCreated).JSON(pagination.NewResponse(new(workspace.WorkspaceResponseDto).Marshall(model)))

}

// Update validate dto , validate user authenticity & update workspace
func Update(c *fiber.Ctx) error {

	model := c.Locals("workspace").(workspace.Workspace)

	dto := new(workspace.UpdateWorkspaceRequestDto)
	dto.ID = model.ID

	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}

	err := workspace.Validate(dto)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	txHandle := sqlclient.Begin()
	err = workspaceService.WithTransaction(txHandle).Update(dto, &model)
	if err != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(err.StatusCode()).JSON(err)
	}
	sqlclient.Commit(txHandle)

	return c.Status(http.StatusOK).JSON(pagination.NewResponse(new(workspace.WorkspaceResponseDto).Marshall(&model)))
}

// Delete deletes user workspace and associated namespace
func Delete(c *fiber.Ctx) error {
	model := c.Locals("workspace").(workspace.Workspace)
	userId := c.Locals("user").(token.UserDetails).ID

	count, err := workspaceService.WithoutTransaction().CountByUserId(userId)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}
	if count == 1 {
		badReq := restErrors.NewBadRequestError("request declined, you should have at least 1 workspace!")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}

	txHandle := sqlclient.Begin()
	err = workspaceService.WithTransaction(txHandle).Delete(&model)
	if err != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(err.StatusCode()).JSON(err)
	}

	err = namespaceService.Delete(model.K8sNamespace)
	if err != nil && err.StatusCode() != http.StatusNotFound {
		sqlclient.Rollback(txHandle)
		return c.Status(err.StatusCode()).JSON(err)
	}

	sqlclient.Commit(txHandle)

	return c.SendStatus(http.StatusNoContent)
}

// GetByUserId find workspaces by userId
func GetByUserId(c *fiber.Ctx) error {
	userId := c.Locals("user").(token.UserDetails).ID

	list, err := workspaceService.WithoutTransaction().GetByUserId(userId)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}
	var marshalled = make([]workspace.WorkspaceResponseDto, 0)
	for _, v := range list {
		record := new(workspace.WorkspaceResponseDto).Marshall(v)
		marshalled = append(marshalled, *record)
	}

	return c.Status(http.StatusOK).JSON(pagination.NewResponse(marshalled))
}

// GetById find workspaces by Id
func GetById(c *fiber.Ctx) error {
	model := c.Locals("workspace").(workspace.Workspace)
	workspaceUser := c.Locals("workspaceUser").(workspaceuser.WorkspaceUser)

	marshalled := new(workspace.WorkspaceResponseDto).Marshall(&model)
	marshalled.Role = workspaceUser.Role

	return c.Status(http.StatusOK).JSON(pagination.NewResponse(marshalled))
}

// AddMember adds new member to workspace
func AddMember(c *fiber.Ctx) error {
	dto := new(workspace.AddWorkspaceMemberDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}

	err := workspace.Validate(dto)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	member, err := userService.WithoutTransaction().GetByEmail(dto.Email)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	model := c.Locals("workspace").(workspace.Workspace)

	for _, v := range model.WorkspaceUsers {
		if v.UserId == member.ID {
			conflictErr := restErrors.NewConflictError("User is already a member of the workspace")
			return c.Status(conflictErr.StatusCode()).JSON(conflictErr)
		}
	}

	txHandle := sqlclient.Begin()
	err = workspaceService.WithTransaction(txHandle).AddWorkspaceMember(&model, member.ID, dto.Role)
	if err != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(err.StatusCode()).JSON(err)
	}

	sqlclient.Commit(txHandle)

	mailRequestDto := new(sendgrid.WorkspaceInvitationMailRequestDto)
	mailRequestDto.Email = dto.Email
	mailRequestDto.WorkspaceName = model.Name
	mailRequestDto.WorkspaceId = model.ID
	go mailService.WorkspaceInvitation(mailRequestDto)

	return c.Status(http.StatusOK).JSON(pagination.NewResponse(pagination.SuccessMessage{
		Message: "user has been added to the workspace",
	}))
}

// Leave removes workspace member from workspace
func Leave(c *fiber.Ctx) error {
	model := c.Locals("workspace").(workspace.Workspace)
	workspaceUser := c.Locals("workspaceUser").(workspaceuser.WorkspaceUser)
	userId := c.Locals("user").(token.UserDetails).ID

	if workspaceUser.Role == roles.Admin { //user with admin role can leave workspace only if it has another admin
		canLeave := false
		for _, v := range model.WorkspaceUsers {
			if v.UserId != userId && v.Role == roles.Admin {
				canLeave = true
				break
			}
		}

		if !canLeave {
			err := restErrors.NewForbiddenError("user with admin role can leave workspace only if it has another admin")
			return c.Status(err.StatusCode()).JSON(err)
		}
	}

	txHandle := sqlclient.Begin()
	err := workspaceService.WithTransaction(txHandle).DeleteWorkspaceMember(&model, userId)
	if err != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(err.StatusCode()).JSON(err)
	}

	sqlclient.Commit(txHandle)

	return c.Status(http.StatusOK).JSON(pagination.NewResponse(pagination.SuccessMessage{
		Message: "You're no longer member of this workspace",
	}))
}

// RemoveMember workspace owner removes workspace member form his/her workspace
func RemoveMember(c *fiber.Ctx) error {
	model := c.Locals("workspace").(workspace.Workspace)
	memberId := c.Params("user_id")
	userId := c.Locals("user").(token.UserDetails).ID

	if memberId == userId {
		badReq := restErrors.NewBadRequestError("you can't remove your self, try to leave workspace instead!")
		return c.Status(badReq.StatusCode()).JSON(badReq)
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
		return c.Status(notFoundErr.StatusCode()).JSON(notFoundErr)
	}

	txHandle := sqlclient.Begin()
	err := workspaceService.WithTransaction(txHandle).DeleteWorkspaceMember(&model, memberId)
	if err != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(err.StatusCode()).JSON(err)
	}

	sqlclient.Commit(txHandle)

	return c.Status(http.StatusOK).JSON(pagination.NewResponse(pagination.SuccessMessage{
		Message: "User has been removed from workspace",
	}))
}

// Members returns a list of workspace members
func Members(c *fiber.Ctx) error {
	model := c.Locals("workspace").(workspace.Workspace)
	userIds := make([]string, len(model.WorkspaceUsers))
	for k, v := range model.WorkspaceUsers {
		userIds[k] = v.UserId
	}

	workspaceMembersList, err := userService.WithoutTransaction().FindWhereIdInSlice(userIds)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	result := make([]user.PublicUserResponseDto, len(workspaceMembersList))
	for k, v := range workspaceMembersList {
		//marshal user model to public response dto
		result[k] = new(user.PublicUserResponseDto).Marshall(v)
		//assign user role
		for _, workspaceUser := range model.WorkspaceUsers {
			if v.ID == workspaceUser.UserId {
				result[k].Role = workspaceUser.Role
			}
		}
	}

	return c.Status(http.StatusOK).JSON(pagination.NewResponse(result))
}

// UpdateWorkspaceUser enables user to assign specific
func UpdateWorkspaceUser(c *fiber.Ctx) error {
	//request dto validation
	dto := new(workspace.UpdateWorkspaceUserRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}
	err := workspace.Validate(dto)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	model := c.Locals("workspace").(workspace.Workspace)
	workspaceUserId := c.Params("user_id")
	userId := c.Locals("user").(token.UserDetails).ID

	//check if the to-be-changed user exists in the workspace
	exist := false
	var workspaceUser *workspaceuser.WorkspaceUser
	for _, v := range model.WorkspaceUsers {
		if v.UserId == workspaceUserId {
			exist = true
			workspaceUser = &v
			break
		}
	}
	if !exist {
		notFoundErr := restErrors.NewNotFoundError("user isn't a member of the workspace")
		return c.Status(notFoundErr.StatusCode()).JSON(notFoundErr)
	}

	if dto.Role != "" { //update workspace-user record role
		if workspaceUser.UserId == userId {
			badReq := restErrors.NewBadRequestError("users can't change their own workspace role!")
			return c.Status(badReq.StatusCode()).JSON(badReq)
		}

	}

	txHandle := sqlclient.Begin()
	err = workspaceService.WithTransaction(txHandle).UpdateWorkspaceUser(workspaceUser, dto)
	if err != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(err.StatusCode()).JSON(err)
	}

	sqlclient.Commit(txHandle)

	return c.Status(http.StatusOK).JSON(pagination.NewResponse(pagination.SuccessMessage{
		Message: "User role changed successfully",
	}))
}

// ValidateWorkspaceExist checks if the workspace_id (id) which passed as query param exist, creates workspace local
func ValidateWorkspaceExist(c *fiber.Ctx) error {
	workspaceId := c.Params("id")

	model, err := workspaceService.WithoutTransaction().GetById(workspaceId)
	if err != nil {
		if err.StatusCode() == http.StatusNotFound {
			notFoundErr := restErrors.NewNotFoundError("no such record")
			return c.Status(notFoundErr.StatusCode()).JSON(notFoundErr)
		}
		return c.Status(err.StatusCode()).JSON(err)
	}

	c.Locals("workspace", *model)

	return c.Next()
}
