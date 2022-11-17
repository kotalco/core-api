package sendgrid

import (
	"github.com/kotalco/cloud-api/pkg/config"
	k8svc "github.com/kotalco/cloud-api/pkg/k8s/svc"
	"github.com/kotalco/cloud-api/pkg/keystore/dbkeystore"
	"github.com/kotalco/community-api/pkg/logger"
)

type MailRequestDto struct {
	Email string
	Token string
}

type WorkspaceInvitationMailRequestDto struct {
	Email         string
	WorkspaceName string
	WorkspaceId   string
}

func GetDomainBaseUrl() string {
	dbKeystore := dbkeystore.NewService()
	url, _ := dbKeystore.Get(config.Environment.DomainMatchBaseURL)
	if url != "" {
		return url
	}

	k8service := k8svc.NewService()
	record, err := k8service.Get("traefik", "traefik")
	if err != nil {
		go logger.Error("SEND_GRID_GET_DOMAIN_BASE_URL", err)
		return ""
	}
	return record.Status.LoadBalancer.Ingress[0].IP

}
