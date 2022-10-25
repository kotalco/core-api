package svc

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/internal/workspace"
	k8svc "github.com/kotalco/cloud-api/pkg/k8s/svc"
	"github.com/kotalco/community-api/pkg/shared"
	"net/http"
)

var (
	svcService        = k8svc.NewService()
	availableProtocol = k8svc.AvailableProtocol
)

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
		ApiEnabled := false
		for _, port := range corv1Service.Spec.Ports { // iterate over service ports
			if availableProtocol(port.Name) { // check if service has API enabled (valid protocols)
				ApiEnabled = true
			}
		}
		if ApiEnabled == true {
			responseDto = append(responseDto, k8svc.SvcDto{Name: corv1Service.Name})
		}
	}

	return c.Status(http.StatusOK).JSON(shared.NewResponse(responseDto))
}
