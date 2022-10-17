package endpoint

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/internal/endpoint"
	"github.com/kotalco/cloud-api/pkg/k8s"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/shared"
	"net/http"
)

var endpointService = k8s.NewEndpointService()

//Todo add user security checks

func Create(c *fiber.Ctx) error {
	dto := new(endpoint.CreateEndpointDto)
	if intErr := c.BodyParser(dto); intErr != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.Status).JSON(badReq)
	}

	err := endpoint.Validate(dto)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	err = endpointService.Create(dto.Name, dto.Namespace, dto.ServiceName, dto.ServicePort)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	return c.Status(http.StatusCreated).JSON(shared.NewResponse(shared.SuccessMessage{
		Message: "Ingress route created",
	}))
}

func List(c *fiber.Ctx) error {
	var namespace = c.Query("namespace", "default")
	list, err := endpointService.List(namespace)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	return c.Status(http.StatusOK).JSON(shared.NewResponse(list))
}

func Get(c *fiber.Ctx) error {
	var namespace = c.Query("namespace", "default")
	var name = c.Params("name")
	record, err := endpointService.Get(name, namespace)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	return c.Status(http.StatusOK).JSON(shared.NewResponse(record))
}
