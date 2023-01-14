package setting

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/internal/setting"
	"github.com/kotalco/cloud-api/pkg/k8s/ingressroute"
	k8svc "github.com/kotalco/cloud-api/pkg/k8s/svc"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	"github.com/kotalco/community-api/pkg/shared"
	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	"net"
	"net/http"
)

var (
	settingService      = setting.NewService()
	k8service           = k8svc.NewService()
	ingressRouteService = ingressroute.NewIngressRoutesService()
)

func VerificationTxtRecord(c *fiber.Ctx) error {
	value, err := settingService.GetDomainTxtRecord()
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	return c.Status(http.StatusOK).JSON(shared.NewResponse(setting.DomainVerificationTxtRecordResponseDto{DomainVerificationTxtRecord: value}))
}

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

	//validate txt record
	dnsTxtRecords, lookErr := net.LookupTXT(dto.Domain)
	if lookErr != nil {
		go logger.Error("CONFIGURE_DOMAIN", lookErr)
		badReq := restErrors.NewBadRequestError(lookErr.Error())
		return c.Status(badReq.Status).JSON(badReq)
	}
	txtRecord, err := settingService.GetDomainTxtRecord()
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	verified := false
	for _, record := range dnsTxtRecords {
		if record == txtRecord {
			verified = true
		}
	}
	if !verified {
		badReqErr := restErrors.NewBadRequestError("can't find kotal-verification-txt-record")
		return c.Status(badReqErr.Status).JSON(badReqErr)
	}

	txHandle := sqlclient.Begin()
	err = settingService.WithTransaction(txHandle).ConfigureDomain(dto)
	if err != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(err.Status).JSON(err)
	}

	//Update API and dashboard ingress routes
	//get ingressRoute
	kotalStackIR, err := ingressRouteService.Get("kotal-stack", "kotal")
	if err != nil {
		go logger.Warn("CONFIGURE_DOMAIN", err)
		sqlclient.Rollback(txHandle)
		return c.Status(err.Status).JSON(err)
	}

	//update ingress-route
	kotalStackIR.Spec.TLS = &traefikv1alpha1.TLS{
		CertResolver: "myresolver",
	}
	kotalStackIR.Spec.EntryPoints = []string{"websecure"}

	for i, v := range kotalStackIR.Spec.Routes {
		switch v.Services[0].Name {
		case "kotal-dashboard":
			kotalStackIR.Spec.Routes[i].Match = fmt.Sprintf("Host(`%s`) && PathPrefix(`/`)", dto.Domain)
		case "kotal-api":
			kotalStackIR.Spec.Routes[i].Match = fmt.Sprintf("Host(`%s`) && PathPrefix(`/api`)", dto.Domain)
		}
	}

	err = ingressRouteService.Update(kotalStackIR)
	if err != nil {
		go logger.Warn("CONFIGURE_DOMAIN", err)
		sqlclient.Rollback(txHandle)
		return c.Status(err.Status).JSON(err)
	}

	sqlclient.Commit(txHandle)
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
