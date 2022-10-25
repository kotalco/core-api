package svc

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/internal/workspace"
	k8svc "github.com/kotalco/cloud-api/pkg/k8s/svc"
	"github.com/kotalco/community-api/pkg/shared"
	"net/http"
)

var svcService = k8svc.NewService()

// List accept namespace, returns a list of names of corv1.Service
// list returns only service with API enabled
func List(c *fiber.Ctx) error {
	workspaceModel := c.Locals("workspace").(workspace.Workspace)

	svcList, err := svcService.List(workspaceModel.K8sNamespace)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	responseDto := make([]k8svc.SvcDto, 0)
	for _, corv1Service := range svcList.Items { // iterate over corv1.ServiceList to create svcDto
		for _, port := range corv1Service.Spec.Ports { // iterate over service ports
			if k8svc.AvailableProtocol(port.Name) { // check if service has API enabled (valid protocols)
				responseDto = append(responseDto, k8svc.SvcDto{Name: port.Name})
			}
		}
	}

	return c.Status(http.StatusOK).JSON(shared.NewResponse(responseDto))
}