package setting

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/core-api/core/setting"
	"github.com/kotalco/core-api/k8s/ingressroute"
	k8svc "github.com/kotalco/core-api/k8s/svc"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/logger"
	"github.com/kotalco/core-api/pkg/responder"
	"github.com/kotalco/core-api/pkg/sqlclient"
	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	"net"
	"net/http"
)

var (
	settingService      = setting.NewService()
	k8service           = k8svc.NewService()
	ingressRouteService = ingressroute.NewIngressRoutesService()
)

func ConfigureDomain(c *fiber.Ctx) error {
	dto := new(setting.ConfigureDomainRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}

	restErr := setting.Validate(dto)
	if restErr != nil {
		return c.Status(restErr.StatusCode()).JSON(restErr)
	}

	ip, hostName, err := networkIdentifiers()
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	if ip != "" {
		err = verifyDomainIP(dto.Domain, ip)
		if err != nil {
			return c.Status(err.StatusCode()).JSON(err)
		}
	} else if hostName != "" {
		err = verifyDomainHostName(dto.Domain, hostName)
		if err != nil {
			return c.Status(err.StatusCode()).JSON(err)
		}
	}

	txHandle := sqlclient.Begin()
	err = settingService.WithTransaction(txHandle).ConfigureDomain(dto)
	if err != nil {
		sqlclient.Rollback(txHandle)
		return c.Status(err.StatusCode()).JSON(err)
	}

	//Update API and dashboard ingress routes
	//get ingressRoute
	kotalStackIR, err := ingressRouteService.Get("kotal-stack", "kotal")
	if err != nil {
		go logger.Warn("CONFIGURE_DOMAIN", err)
		sqlclient.Rollback(txHandle)
		return c.Status(err.StatusCode()).JSON(err)
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
		return c.Status(err.StatusCode()).JSON(err)
	}

	sqlclient.Commit(txHandle)
	return c.Status(http.StatusOK).JSON(responder.NewResponse(responder.SuccessMessage{Message: "domain configured successfully!"}))
}

func ConfigureRegistration(c *fiber.Ctx) error {
	dto := new(setting.ConfigureRegistrationRequestDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}
	restErr := setting.Validate(dto)
	if restErr != nil {
		return c.Status(restErr.StatusCode()).JSON(restErr)
	}

	err := settingService.WithoutTransaction().ConfigureRegistration(dto)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}
	return c.Status(http.StatusOK).JSON(responder.NewResponse(responder.SuccessMessage{Message: "registration configured successfully!"}))
}

func Settings(c *fiber.Ctx) error {
	list, err := settingService.WithoutTransaction().Settings()
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	marshalledList := make([]setting.SettingResponseDto, 0)
	for _, v := range list {
		marshalledList = append(marshalledList, new(setting.SettingResponseDto).Marshall(v))
	}
	return c.Status(http.StatusOK).JSON(responder.NewResponse(marshalledList))
}

func NetworkIdentifiers(c *fiber.Ctx) error {
	ipAddress, hostName, err := networkIdentifiers()
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}
	return c.Status(http.StatusOK).JSON(responder.NewResponse(&setting.NetworkIdentifierResponseDto{IPAddress: ipAddress, HostName: hostName}))
}

var networkIdentifiers = func() (ip string, hostName string, restErr restErrors.IRestErr) {
	record, restErr := k8service.Get("traefik", "traefik")
	if restErr != nil {
		if restErr.StatusCode() == http.StatusNotFound {
			record, restErr = k8service.Get("kotal-traefik", "traefik")
			if restErr != nil {
				go logger.Error("SETTING_GET_IP_ADDRESS", restErr)
				return
			}
		} else {
			go logger.Error("SETTING_GET_IP_ADDRESS", restErr)
			return
		}
	}

	defer func() {
		if err := recover(); err != nil {
			restErr = restErrors.NewNotFoundError("can't get network identifiers, still provisioning!")
			return
		}
	}()
	ip = record.Status.LoadBalancer.Ingress[0].IP
	hostName = record.Status.LoadBalancer.Ingress[0].Hostname

	return
}

var verifyDomainIP = func(domain string, ipAddress string) restErrors.IRestErr {
	records, lookErr := net.LookupHost(domain)
	if lookErr != nil {
		go logger.Error("VERIFY_DOMAIN_A_RECORDS", lookErr)
		badReq := restErrors.NewBadRequestError(lookErr.Error())
		return badReq
	}
	verified := false
	for _, record := range records {
		if record == ipAddress {
			verified = true
		}
	}
	if !verified {
		badReqErr := restErrors.NewBadRequestError(fmt.Sprintf("Domain DNS records hasn't been updated with an A record that maps %s to %s.", domain, ipAddress))
		return badReqErr
	}
	return nil
}

var verifyDomainHostName = func(domain string, hostName string) restErrors.IRestErr {
	CName, lookErr := net.LookupCNAME(domain)
	if lookErr != nil {
		go logger.Error("VERIFY_DOMAIN_C_Name", lookErr)
		badReq := restErrors.NewBadRequestError(lookErr.Error())
		return badReq
	}

	if CName != hostName {
		badReqErr := restErrors.NewBadRequestError(fmt.Sprintf("Domain hasn't been updated with a CName that maps %s to %s.", domain, hostName))
		return badReqErr
	}
	return nil
}
