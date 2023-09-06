package endpoint

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/core/endpoint"
	"github.com/kotalco/cloud-api/core/endpointactivity"
	"github.com/kotalco/cloud-api/core/setting"
	"github.com/kotalco/cloud-api/core/workspace"
	"github.com/kotalco/cloud-api/pkg/k8s/secret"
	k8svc "github.com/kotalco/cloud-api/pkg/k8s/svc"
	"github.com/kotalco/cloud-api/pkg/token"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	"github.com/kotalco/community-api/pkg/shared"
	"net/http"
)

var (
	endpointService   = endpoint.NewService()
	svcService        = k8svc.NewService()
	availableProtocol = k8svc.AvailableProtocol
	settingService    = setting.NewService()
	secretService     = secret.NewService()
	activityService   = endpointactivity.NewService()
)

// Create accept  endpoint.CreateEndpointDto , creates the endpoint and returns success or err if any
func Create(c *fiber.Ctx) error {
	workspaceModel := c.Locals("workspace").(workspace.Workspace)
	userId := c.Locals("user").(token.UserDetails).ID
	dto := new(endpoint.CreateEndpointDto)
	if intErr := c.BodyParser(dto); intErr != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}

	err := endpoint.Validate(dto)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	//check if the user configured the domain base url
	if !settingService.WithoutTransaction().IsDomainConfigured() {
		forbiddenRes := restErrors.NewForbiddenError("Domain hasn't been configured yet !")
		return c.Status(forbiddenRes.StatusCode()).JSON(forbiddenRes)
	}
	//get service
	corev1Svc, err := svcService.Get(dto.ServiceName, workspaceModel.K8sNamespace)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	//check if service has API enabled
	validProtocol := false
	for _, v := range corev1Svc.Spec.Ports {
		if availableProtocol(v.Name) {
			validProtocol = true
		}
	}
	if validProtocol == false {
		badReq := restErrors.NewBadRequestError(fmt.Sprintf("service %s doesn't have API enabled", corev1Svc.Name))
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}

	dto.UserId = userId
	err = endpointService.Create(dto, corev1Svc)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}
	return c.Status(http.StatusCreated).JSON(shared.NewResponse(shared.SuccessMessage{
		Message: "Endpoint has been created",
	}))
}

// List accept namespace , returns a list of ingressroute.Ingressroute  or err if any
func List(c *fiber.Ctx) error {
	workspaceModel := c.Locals("workspace").(workspace.Workspace)
	list, err := endpointService.List(workspaceModel.K8sNamespace, map[string]string{"app.kubernetes.io/created-by": "kotal-api"})
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	marshalledDto := make([]*endpoint.EndpointMetaDto, 0)
	for _, item := range list.Items {
		marshalledDto = append(marshalledDto, new(endpoint.EndpointMetaDto).Marshall(&item))
	}

	c.Set("Access-Control-Expose-Headers", "X-Total-Count")
	c.Set("X-Total-Count", fmt.Sprintf("%d", len(marshalledDto)))

	return c.Status(http.StatusOK).JSON(shared.NewResponse(marshalledDto))
}

// Get accept namespace and name , returns a record of type ingressroute.Ingressroute or err if any
func Get(c *fiber.Ctx) error {
	workspaceModel := c.Locals("workspace").(workspace.Workspace)
	endpointName := c.Params("name")

	record, err := endpointService.Get(endpointName, workspaceModel.K8sNamespace)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	//get secret
	secretName := fmt.Sprintf("%s-secret", record.Name)
	v1Secret, _ := secretService.Get(secretName, workspaceModel.K8sNamespace)
	endpointDto := new(endpoint.EndpointDto).Marshall(record, v1Secret)

	return c.Status(http.StatusOK).JSON(shared.NewResponse(endpointDto))
}

// Delete accept namespace and the name of the ingress-route ,deletes it , returns success message or err if any
func Delete(c *fiber.Ctx) error {
	workspaceModel := c.Locals("workspace").(workspace.Workspace)
	endpointName := c.Params("name")

	err := endpointService.Delete(endpointName, workspaceModel.K8sNamespace)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}
	return c.Status(http.StatusOK).JSON(shared.NewResponse(shared.SuccessMessage{Message: "Endpoint Deleted"}))
}

func Count(c *fiber.Ctx) error {
	workspaceModel := c.Locals("workspace").(workspace.Workspace)

	count, err := endpointService.Count(workspaceModel.K8sNamespace, map[string]string{"app.kubernetes.io/created-by": "kotal-api"})
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	c.Set("Access-Control-Expose-Headers", "X-Total-Count")
	c.Set("X-Total-Count", fmt.Sprintf("%d", count))

	return c.SendStatus(http.StatusOK)
}

func WriteStats(c *fiber.Ctx) error {
	dto := new(endpointactivity.CreateEndpointActivityDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		go logger.Error("ENDPOINT_ACTIVITY_HANDLER_WRITE_STATS", err)
		return c.SendStatus(badReq.StatusCode())
	}

	err := endpointactivity.Validate(dto)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	err = activityService.Create(dto.RequestId)
	if err != nil {
		go logger.Error("ENDPOINT_ACTIVITY_HANDLER_WRITE_STATS", err)
		return c.SendStatus(err.StatusCode())
	}
	return c.SendStatus(http.StatusOK)
}

func ReadStats(c *fiber.Ctx) error {
	workspaceModel := c.Locals("workspace").(workspace.Workspace)
	endpointName := c.Params("name")

	record, err := endpointService.Get(endpointName, workspaceModel.K8sNamespace)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	dto := map[string]endpointactivity.ActivityAggregations{}
	for _, v := range record.Spec.Routes {
		portName := v.Services[0].Port.StrVal
		if k8svc.AvailableProtocol(portName) {
			monthlyCount, err := activityService.MonthlyActivity(endpointactivity.GetEndpointId(v.Match))
			if err != nil {
				go logger.Error("ENDPOINT_ACTIVITY_HANDLER_READ_STATS", err)
				return c.Status(err.StatusCode()).JSON(err)
			}
			dto[portName] = endpointactivity.ActivityAggregations{MonthlyHits: *monthlyCount}
		}

	}

	return c.Status(http.StatusOK).JSON(shared.NewResponse(dto))
}
