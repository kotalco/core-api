package svc

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/core/workspace"
	"github.com/kotalco/cloud-api/k8s/svc"
	"github.com/kotalco/cloud-api/pkg/pagination"
	"net/http"
)

var (
	svcService        = svc.NewService()
	availableProtocol = svc.AvailableProtocol
)

// List accept namespace, returns a list of names of corv1.Service
// list returns only service with API enabled
func List(c *fiber.Ctx) error {
	workspaceModel := c.Locals("workspace").(workspace.Workspace)

	svcList, err := svcService.List(workspaceModel.K8sNamespace)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	responseDto := make([]svc.SvcDto, 0)
	for _, corv1Service := range svcList.Items { // iterate over corv1.ServiceList to create svcDto
		ApiEnabled := false
		for _, port := range corv1Service.Spec.Ports { // iterate over service ports
			if availableProtocol(port.Name) { // check if service has API enabled (valid protocols)
				ApiEnabled = true
			}
		}
		if ApiEnabled == true {
			responseDto = append(responseDto, svc.SvcDto{Name: corv1Service.Name, Protocol: corv1Service.Labels["kotal.io/protocol"]})
		}
	}

	return c.Status(http.StatusOK).JSON(pagination.NewResponse(responseDto))
}
