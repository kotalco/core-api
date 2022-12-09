package setting

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/internal/setting"
	k8svc "github.com/kotalco/cloud-api/pkg/k8s/svc"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	"github.com/kotalco/community-api/pkg/shared"
	"net/http"
)

var (
	settingService = setting.NewService()
	k8service      = k8svc.NewService()
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

	err := settingService.WithoutTransaction().ConfigureDomain(dto)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	//set domain baseUrl env variable
	return c.Status(http.StatusOK).JSON(shared.NewResponse(shared.SuccessMessage{Message: "domain configured successfully!"}))
}

func ConfigureRegistration(c *fiber.Ctx) error {
	dto := new(setting.ConfigureRegistrationRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.Status).JSON(badReq)
	}
	restErr := setting.Validate(dto)
	if restErr != nil {
		return c.Status(restErr.Status).JSON(restErr)
	}

	err := settingService.WithoutTransaction().ConfigureRegistration(dto)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	return c.Status(http.StatusOK).JSON(shared.NewResponse(shared.SuccessMessage{Message: "registration configured successfully!"}))
}

func Settings(c *fiber.Ctx) error {
	list, err := settingService.WithoutTransaction().Settings()
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	marshalledList := make([]setting.SettingResponseDto, 0)
	for _, v := range list {
		marshalledList = append(marshalledList, new(setting.SettingResponseDto).Marshall(v))
	}
	return c.Status(http.StatusOK).JSON(shared.NewResponse(marshalledList))
}

func IPAddress(c *fiber.Ctx) error {
	record, err := k8service.Get("traefik", "traefik")
	if err != nil {
		if err.Status == http.StatusNotFound {
			record, err = k8service.Get("kotal-traefik", "traefik")
			if err != nil {
				go logger.Error("SETTING_GET_IP_ADDRESS", err)
				return c.Status(err.Status).JSON(err)
			}
		} else {
			go logger.Error("SETTING_GET_IP_ADDRESS", err)
			return c.Status(err.Status).JSON(err)
		}
	}

	defer func() {
		if err := recover(); err != nil {
			c.Status(http.StatusNotFound).JSON(restErrors.NewNotFoundError("can't get ip address, still provisioning!"))
		}
	}()

	return c.Status(http.StatusOK).JSON(shared.NewResponse(&setting.IPAddressResponseDto{IPAddress: record.Status.LoadBalancer.Ingress[0].IP}))
}
