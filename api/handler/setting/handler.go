package setting

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/core-api/config"
	"github.com/kotalco/core-api/core/setting"
	"github.com/kotalco/core-api/core/user"
	"github.com/kotalco/core-api/k8s"
	"github.com/kotalco/core-api/k8s/ingressroute"
	"github.com/kotalco/core-api/k8s/secret"
	k8svc "github.com/kotalco/core-api/k8s/svc"
	"github.com/kotalco/core-api/k8s/tlscertificate"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/logger"
	"github.com/kotalco/core-api/pkg/responder"
	"github.com/kotalco/core-api/pkg/security"
	"github.com/kotalco/core-api/pkg/sendgrid"
	"github.com/kotalco/core-api/pkg/sqlclient"
	"github.com/kotalco/core-api/pkg/token"
	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net"
	"net/http"
)

var (
	settingService        = setting.NewService()
	k8service             = k8svc.NewService()
	ingressRouteService   = ingressroute.NewIngressRoutesService()
	secretService         = secret.NewService()
	tlsCertificateService = tlscertificate.NewService()
	userService           = user.NewService()
	sendGridService       = sendgrid.NewService()
	k8sClient             = k8s.NewClientService()
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
	kotalStackIR, err := ingressRouteService.Get(config.Environment.KotalIngressRouteName, config.Environment.KotalNamespace)
	if err != nil {
		go logger.Warn("CONFIGURE_DOMAIN", err)
		sqlclient.Rollback(txHandle)
		return c.Status(err.StatusCode()).JSON(err)
	}

	//update ingress-route
	kotalStackIR.Spec.TLS = &traefikv1alpha1.TLS{
		CertResolver: setting.KotalLetsEncryptResolverName,
	}
	kotalStackIR.Spec.EntryPoints = []string{"websecure"}

	for i, v := range kotalStackIR.Spec.Routes {
		switch v.Services[0].Name {
		case "kotal-dashboard":
			kotalStackIR.Spec.Routes[i].Match = fmt.Sprintf("Host(`app.%s`) && PathPrefix(`/`)", dto.Domain)
		case "kotal-api":
			kotalStackIR.Spec.Routes[i].Match = fmt.Sprintf("Host(`app.%s`) && PathPrefix(`/api`)", dto.Domain)
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

	err := sendGridService.Ping()
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	err = settingService.WithoutTransaction().ConfigureRegistration(dto)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}
	return c.Status(http.StatusOK).JSON(responder.NewResponse(responder.SuccessMessage{Message: "registration configured successfully!"}))
}

func ConfigureTLS(c *fiber.Ctx) error {
	userId := c.Locals("user").(token.UserDetails).ID
	userDetails, restErr := userService.WithoutTransaction().GetById(userId)
	if restErr != nil {
		return c.Status(restErr.StatusCode()).JSON(restErr)
	}
	switch c.FormValue("tls_provider") {
	case "letsencrypt":
		//Sets LetsEncrypt static Configuration
		restErr = tlsCertificateService.ConfigureLetsEncrypt(setting.KotalLetsEncryptResolverName, userDetails.Email)
		if restErr != nil {
			logger.Error("CONFIGURE_TLS", restErr)
			return c.Status(restErr.StatusCode()).JSON(restErr)
		}
	case "secret":
		fileHeaderCert, err := c.FormFile("cert")
		if err != nil {
			badReq := restErrors.NewBadRequestError("missing cert file")
			return c.Status(badReq.StatusCode()).JSON(badReq)
		}
		fileHeaderKey, err := c.FormFile("key")
		if err != nil {
			badReq := restErrors.NewBadRequestError("missing key file")
			return c.Status(badReq.StatusCode()).JSON(badReq)
		}

		certBytes, err := security.ValidateFile(fileHeaderCert)
		if err != nil {
			badReq := restErrors.NewBadRequestError(err.Error())
			return c.Status(badReq.StatusCode()).JSON(badReq)
		}

		keyBytes, err := security.ValidateFile(fileHeaderKey)
		if err != nil {
			badReq := restErrors.NewBadRequestError(err.Error())
			return c.Status(badReq.StatusCode()).JSON(badReq)
		}

		_ = secretService.Delete(setting.CustomTLSSecretName, config.Environment.KotalNamespace)

		restErr = secretService.Create(&secret.CreateSecretDto{
			ObjectMeta: metav1.ObjectMeta{
				Name:      setting.CustomTLSSecretName,
				Namespace: config.Environment.KotalNamespace,
			},
			Type: corev1.SecretTypeTLS,
			Data: map[string][]byte{
				"tls.crt": certBytes,
				"tls.key": keyBytes,
			},
		})
		if restErr != nil {
			return c.Status(restErr.StatusCode()).JSON(restErr)
		}
		restErr = tlsCertificateService.ConfigureCustomCertificate(setting.CustomTLSSecretName)
		if restErr != nil {
			return c.Status(restErr.StatusCode()).JSON(restErr)
		}
	default:
		badReq := restErrors.NewBadRequestError("tls_provider can be only be letsencrypt or secret")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}
	return c.Status(http.StatusOK).JSON(responder.SuccessMessage{Message: "tls certificate configured successfully"})
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
