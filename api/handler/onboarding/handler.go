package onboarding

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/internal/onboarding"
	"github.com/kotalco/cloud-api/pkg/sendgrid"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/shared"
	"net/http"
)

var (
	onboardingService = onboarding.NewService()
)

func SetDomain(c *fiber.Ctx) error {

	fmt.Println(sendgrid.GetDomainBaseUrl())
	dto := new(onboarding.SetDomainBaseUrlRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.Status).JSON(badReq)
	}

	restErr := onboarding.Validate(dto)
	if restErr != nil {
		return c.Status(restErr.Status).JSON(restErr)
	}

	err := onboardingService.SetDomainBaseurl(dto)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	return c.Status(http.StatusOK).JSON(shared.NewResponse(shared.SuccessMessage{Message: "domain added successfully!"}))
}
