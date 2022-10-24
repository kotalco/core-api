package svc

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/internal/endpoint"
	"github.com/kotalco/cloud-api/internal/workspace"
	"github.com/kotalco/cloud-api/pkg/svc"
	"github.com/kotalco/community-api/pkg/shared"
	"net/http"
)

var svcService = svc.NewService()

func ListServices(c *fiber.Ctx) error {
	workspaceModel := c.Locals("workspace").(workspace.Workspace)

	svcList, err := svcService.List(workspaceModel.K8sNamespace)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	responseDto := make([]endpoint.SvcResponseDto, 0)
	for _, v := range svcList.Items {
		responseDto = append(responseDto, endpoint.SvcResponseDto{Name: v.Name})
	}
	return c.Status(http.StatusOK).JSON(shared.NewResponse(responseDto))
}
