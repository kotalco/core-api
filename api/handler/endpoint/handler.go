package endpoint

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/core-api/core/endpoint"
	"github.com/kotalco/core-api/core/endpointactivity"
	"github.com/kotalco/core-api/core/setting"
	"github.com/kotalco/core-api/core/workspace"
	"github.com/kotalco/core-api/k8s/secret"
	"github.com/kotalco/core-api/k8s/svc"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/logger"
	"github.com/kotalco/core-api/pkg/pagination"
	"github.com/kotalco/core-api/pkg/responder"
	"github.com/kotalco/core-api/pkg/token"
	"net/http"
	"sort"
	"strconv"
	"time"
)

const ActivityDateLayout = "02-01-2006"

var (
	endpointService   = endpoint.NewService()
	svcService        = svc.NewService()
	availableProtocol = svc.AvailableProtocol
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
	return c.Status(http.StatusCreated).JSON(responder.NewResponse(responder.SuccessMessage{
		Message: "Endpoint has been created",
	}))
}

// List accept namespace , returns a list of ingressroute.Ingressroute  or err if any
func List(c *fiber.Ctx) error {
	// default page to 0
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	workspaceModel := c.Locals("workspace").(workspace.Workspace)

	list, err := endpointService.List(workspaceModel.K8sNamespace, map[string]string{"app.kubernetes.io/created-by": "kotal-api"})
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	start, end := pagination.Page(uint(len(list.Items)), uint(page), uint(limit))
	sort.Slice(list.Items[:], func(i, j int) bool {
		return list.Items[j].CreationTimestamp.Before(&list.Items[i].CreationTimestamp)
	})

	c.Set("Access-Control-Expose-Headers", "X-Total-Count")
	c.Set("X-Total-Count", fmt.Sprintf("%d", len(list.Items)))

	marshalledDto := make([]*endpoint.EndpointMetaDto, 0)
	for _, item := range list.Items[start:end] {
		marshalledDto = append(marshalledDto, new(endpoint.EndpointMetaDto).Marshall(&item))
	}

	return c.Status(http.StatusOK).JSON(responder.NewResponse(marshalledDto))
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

	return c.Status(http.StatusOK).JSON(responder.NewResponse(endpointDto))
}

// Delete accept namespace and the name of the ingress-route ,deletes it , returns success message or err if any
func Delete(c *fiber.Ctx) error {
	workspaceModel := c.Locals("workspace").(workspace.Workspace)
	endpointName := c.Params("name")

	err := endpointService.Delete(endpointName, workspaceModel.K8sNamespace)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}
	return c.Status(http.StatusOK).JSON(responder.NewResponse(responder.SuccessMessage{Message: "Endpoint Deleted"}))
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
	var dtos []endpointactivity.CreateEndpointActivityDto
	if err := c.BodyParser(&dtos); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		go logger.Error("ENDPOINT_ACTIVITY_HANDLER_WRITE_STATS", err)
		return c.SendStatus(badReq.StatusCode())
	}

	for _, dto := range dtos {
		err := endpointactivity.Validate(dto)
		if err != nil {
			return c.Status(err.StatusCode()).JSON(err)
		}
	}

	err := activityService.Create(dtos)
	if err != nil {
		go logger.Error("ENDPOINT_ACTIVITY_HANDLER_WRITE_STATS", err)
		return c.SendStatus(err.StatusCode())
	}
	return c.SendStatus(http.StatusOK)
}

func ReadStats(c *fiber.Ctx) error {
	workspaceModel := c.Locals("workspace").(workspace.Workspace)
	endpointName := c.Params("name")

	statsType := c.Query("type")
	if statsType == "" {
		badReq := restErrors.NewBadRequestError("query parameter 'type' cannot be empty")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}

	var endDate = time.Now()
	var startDate time.Time

	switch statsType {
	case endpointactivity.LastMonth:
		startDate = endDate.AddDate(0, -1, 0)
	case endpointactivity.LastWeek:
		startDate = endDate.AddDate(0, 0, -7)
	default:
		badReq := restErrors.NewBadRequestError(fmt.Sprintf("query parameter 'type' must be '%s' or '%s'", endpointactivity.LastMonth, endpointactivity.LastWeek))
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}

	record, err := endpointService.Get(endpointName, workspaceModel.K8sNamespace)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	dto := map[string][]map[time.Time]int{}
	for _, v := range record.Spec.Routes {
		portName := v.Services[0].Port.StrVal
		if svc.AvailableProtocol(portName) {
			stats := make([]map[time.Time]int, 0)
			activities, _ := activityService.Stats(startDate, endDate, endpointactivity.GetEndpointId(v.Match))
			activitiesMap := make(map[string]int)
			for _, activity := range *activities {
				activitiesMap[activity.Date.Format(time.DateOnly)] = activity.Activity
			}

			for dt := startDate; !dt.After(endDate); dt = dt.AddDate(0, 0, 1) {
				dtString := dt.Format(time.DateOnly)
				activity, ok := activitiesMap[dtString]
				if !ok {
					activity = 0
				}
				stats = append(stats, map[time.Time]int{dt: activity})
			}

			dto[portName] = stats
		}
	}

	return c.Status(http.StatusOK).JSON(responder.NewResponse(dto))
}
