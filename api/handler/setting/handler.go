package setting

import (
	"crypto/tls"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/core-api/config"
	"github.com/kotalco/core-api/core/setting"
	"github.com/kotalco/core-api/k8s/deployment"
	"github.com/kotalco/core-api/k8s/ingressroute"
	"github.com/kotalco/core-api/k8s/secret"
	k8svc "github.com/kotalco/core-api/k8s/svc"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/logger"
	"github.com/kotalco/core-api/pkg/responder"
	"github.com/kotalco/core-api/pkg/sqlclient"
	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"net"
	"net/http"
	"strings"
)

var (
	settingService      = setting.NewService()
	k8service           = k8svc.NewService()
	ingressRouteService = ingressroute.NewIngressRoutesService()
	secretService       = secret.NewService()
	deploymentService   = deployment.NewService()
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
		CertResolver: config.Environment.LetsEncryptResolverName,
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

	err := settingService.WithoutTransaction().ConfigureRegistration(dto)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}
	return c.Status(http.StatusOK).JSON(responder.NewResponse(responder.SuccessMessage{Message: "registration configured successfully!"}))
}

func ConfigureTLS(c *fiber.Ctx) error {
	//get ingressRoute
	kotalStackIR, err := ingressRouteService.Get(config.Environment.KotalIngressRouteName, config.Environment.KotalNamespace)
	if err != nil {
		go logger.Warn("CONFIGURE_TLS", err)
		return c.Status(err.StatusCode()).JSON(err)
	}

	//get deployment
	traefikDep, err := deploymentService.Get(types.NamespacedName{Name: config.Environment.TraefikDeploymentName, Namespace: config.Environment.TraefikNamespace})
	if err != nil {
		go logger.Warn("CONFIGURE_TLS", err)
		return c.Status(err.StatusCode()).JSON(err)
	}

	//remove certificate static configuration if exit
	for i, container := range traefikDep.Spec.Template.Spec.Containers {
		if container.Name == config.Environment.TraefikDeploymentName {
			var newArgs []string
			for _, arg := range container.Args {
				if !strings.Contains(arg, config.Environment.LetsEncryptResolverName) {
					newArgs = append(newArgs, arg)
				}
			}
			traefikDep.Spec.Template.Spec.Containers[i].Args = newArgs
			break
		}
	}

	tlsType := c.FormValue("tls_type")
	switch tlsType {
	case "letsencrypt":
		kotalStackIR.Spec.TLS = &traefikv1alpha1.TLS{
			CertResolver: config.Environment.LetsEncryptResolverName,
		}
		for i, container := range traefikDep.Spec.Template.Spec.Containers {
			if container.Name == config.Environment.TraefikDeploymentName {
				traefikDep.Spec.Template.Spec.Containers[i].Args = append(traefikDep.Spec.Template.Spec.Containers[i].Args, config.Environment.LetsEncryptStaticConfiguration...)
				break
			}
		}
	case "secret":
		//Get Files
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

		//Open Files
		certFile, err := fileHeaderCert.Open()
		if err != nil {
			badReq := restErrors.NewBadRequestError("couldn't open cert file")
			return c.Status(badReq.StatusCode()).JSON(badReq)
		}
		defer certFile.Close()
		keyFile, err := fileHeaderKey.Open()
		if err != nil {
			badReq := restErrors.NewBadRequestError("couldn't open key file")
			return c.Status(badReq.StatusCode()).JSON(badReq)
		}
		defer keyFile.Close()

		//Read Files
		certBytes, err := io.ReadAll(certFile)
		if err != nil {
			badReq := restErrors.NewBadRequestError("couldn't read cert file")
			return c.Status(badReq.StatusCode()).JSON(badReq)
		}
		keyBytes, err := io.ReadAll(keyFile)
		if err != nil {
			badReq := restErrors.NewBadRequestError("couldn't read key file")
			return c.Status(badReq.StatusCode()).JSON(badReq)
		}

		//validate tls files
		_, err = tls.X509KeyPair(certBytes, keyBytes)
		if err != nil {
			badReq := restErrors.NewBadRequestError(fmt.Sprintf("invalid key cert pair: %s", err.Error()))
			return c.Status(badReq.StatusCode()).JSON(badReq)
		}

		//delete the old secret
		_ = secretService.Delete(setting.CustomTLS, config.Environment.KotalNamespace)
		//create secret
		restErr := secretService.Create(&secret.CreateSecretDto{
			ObjectMeta: metav1.ObjectMeta{
				Name:      setting.CustomTLS,
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

		//update ingress-route
		kotalStackIR.Spec.TLS = &traefikv1alpha1.TLS{
			SecretName: setting.CustomTLS,
		}
	default:
		badReq := restErrors.NewBadRequestError("invalid tls_type can be only be letsencrypt or secret")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}

	err = ingressRouteService.Update(kotalStackIR)
	if err != nil {
		go logger.Warn("CONFIGURE_TLS", err)
		return c.Status(err.StatusCode()).JSON(err)
	}

	err = deploymentService.Update(traefikDep)
	if err != nil {
		go logger.Warn("CONFIGURE_TLS", err)
		return c.Status(err.StatusCode()).JSON(err)
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
