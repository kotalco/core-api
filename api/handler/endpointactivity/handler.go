package endpointactivity

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kotalco/cloud-api/core/endpointactivity"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	"net/http"
)

var activityService = endpointactivity.NewService()

func Logs(c *fiber.Ctx) error {
	dto := new(endpointactivity.EndpointActivityDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		go logger.Error("ENDPOINT_ACTIVITY_HANDLER_LOGS", err)
		return c.SendStatus(badReq.StatusCode())
	}

	err := endpointactivity.Validate(dto)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	record, err := activityService.GetByEndpointId(dto.RequestId)
	if err != nil {
		if err.StatusCode() == http.StatusNotFound {
			record = new(endpointactivity.Activity)
			record.ID = uuid.NewString()
			record.EndpointId = dto.RequestId
			record.Counter = 0
		}
	}

	err = activityService.Increment(record)
	if err != nil {
		go logger.Error("ENDPOINT_ACTIVITY_HANDLER_LOGS", err)
		return c.SendStatus(err.StatusCode())
	}
	return c.SendStatus(http.StatusOK)
}
