package endpoint

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/internal/endpoint"
	"github.com/kotalco/cloud-api/internal/workspace"
	k8svc "github.com/kotalco/cloud-api/pkg/k8s/svc"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/shared"
	"net/http"
)

var (
	endpointService = endpoint.NewService()
	svcService      = k8svc.NewService()
)

// Create accept  endpoint.CreateEndpointDto , creates the endpoint and returns success or err if any
func Create(c *fiber.Ctx) error {
	workspaceModel := c.Locals("workspace").(workspace.Workspace)

	dto := new(endpoint.CreateEndpointDto)
	if intErr := c.BodyParser(dto); intErr != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.Status).JSON(badReq)
	}

	err := endpoint.Validate(dto)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	//get service
	corev1Svc, err := svcService.Get(dto.ServiceName, workspaceModel.K8sNamespace)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	//check if service has API enabled
	validProtocol := false
	for _, v := range corev1Svc.Spec.Ports {
		if k8svc.AvailableProtocol(v.Name) {
			validProtocol = true
		}
	}
	if validProtocol == false {
		badReq := restErrors.NewBadRequestError(fmt.Sprintf("service %s doesn't have API enabled", corev1Svc.Name))
		return c.Status(badReq.Status).JSON(badReq)
	}

	err = endpointService.Create(dto, corev1Svc, workspaceModel.K8sNamespace)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	return c.Status(http.StatusCreated).JSON(shared.NewResponse(shared.SuccessMessage{
		Message: "Endpoint has been created",
	}))
}

// List accept namespace , returns a list of ingressroute.Ingressroute  or err if any
func List(c *fiber.Ctx) error {
	workspaceModel := c.Locals("workspace").(workspace.Workspace)
	list, err := endpointService.List(workspaceModel.K8sNamespace)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	return c.Status(http.StatusOK).JSON(shared.NewResponse(list))
}

//Get accept namespace and name , returns a record of type ingressroute.Ingressroute or err if any
func Get(c *fiber.Ctx) error {
	workspaceModel := c.Locals("workspace").(workspace.Workspace)
	endpointName := c.Params("name")

	record, err := endpointService.Get(endpointName, workspaceModel.K8sNamespace)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	return c.Status(http.StatusOK).JSON(shared.NewResponse(record))
}

//Delete accept namespace and the name of the ingress-route ,deletes it , returns success message or err if any
func Delete(c *fiber.Ctx) error {
	workspaceModel := c.Locals("workspace").(workspace.Workspace)
	endpointName := c.Params("name")

	err := endpointService.Delete(endpointName, workspaceModel.K8sNamespace)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	return c.Status(http.StatusOK).JSON(shared.NewResponse(shared.SuccessMessage{Message: "Endpoint Deleted"}))
}
