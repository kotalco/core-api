package domain

import (
	"github.com/kotalco/cloud-api/pkg/config"
	k8svc "github.com/kotalco/cloud-api/pkg/k8s/svc"
	"github.com/kotalco/cloud-api/pkg/keystore/dbkeystore"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
)

// GetDomainBaseUrl get the app baseurl and error if any
// If the user configured his/her domain, if not it gets the traefik external ip
func GetDomainBaseUrl() (string, *restErrors.RestErr) {
	dbKeystore := dbkeystore.NewService()
	url, _ := dbKeystore.Get(config.Environment.DomainMatchBaseURLKey)
	if url != "" {
		return url, nil
	}

	k8service := k8svc.NewService()
	record, err := k8service.Get("traefik", "traefik")
	if err != nil {
		go logger.Error("SEND_GRID_GET_DOMAIN_BASE_URL", err)
		return "", restErrors.NewInternalServerError("can't get traefik service")
	}
	return record.Status.LoadBalancer.Ingress[0].IP, nil
}
