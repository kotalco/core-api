package setting

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/internal/setting"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/shared"
	"net/http"
)

var (
	settingService = setting.NewService()
)

func ConfigureDomain(c *fiber.Ctx) error {
	dto := new(setting.ConfigureDomainRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.Status).JSON(badReq)
	}

	restErr := setting.Validate(dto)
	if restErr != nil {
		return c.Status(restErr.Status).JSON(restErr)
	}

	err := settingService.ConfigureDomain(dto)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	return c.Status(http.StatusOK).JSON(shared.NewResponse(shared.SuccessMessage{Message: "domain configured successfully!"}))
}

func UpdateDomainConfiguration(c *fiber.Ctx) error {
	dto := new(setting.ConfigureDomainRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.Status).JSON(badReq)
	}

	restErr := setting.Validate(dto)
	if restErr != nil {
		return c.Status(restErr.Status).JSON(restErr)
	}

	err := settingService.UpdateDomainConfiguration(dto)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	return c.Status(http.StatusOK).JSON(shared.NewResponse(shared.SuccessMessage{Message: "domain configuration updated successfully!"}))
}

func Settings(c *fiber.Ctx) error {
	record, err := settingService.Settings()
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	return c.Status(http.StatusOK).JSON(shared.NewResponse(new(setting.SettingResponseDto).Marshall(record)))
}
