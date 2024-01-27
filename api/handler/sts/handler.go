package sts

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/core-api/k8s/statefulset"
	"github.com/kotalco/core-api/pkg/responder"
	"net/http"
)

var statefulSetService = statefulset.NewService()

func Count(c *fiber.Ctx) error {
	list, err := statefulSetService.List(c.Locals("namespace").(string))
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	result := map[string]uint{}

	for _, item := range list.Items {
		protocol := item.Labels["kotal.io/protocol"]
		result[protocol]++
	}

	return c.Status(http.StatusOK).JSON(responder.NewResponse(result))
}
