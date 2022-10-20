package endpoint

import (
	"fmt"
	"github.com/kotalco/cloud-api/pkg/config"
	"github.com/kotalco/cloud-api/pkg/k8s/ingressroute"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

var (
	ingressRoutesService = ingressroute.NewIngressRoutesService()
)

type service struct{}

type IService interface {
	Create(dto *CreateEndpointDto, svc *corev1.Service, namespace string) *restErrors.RestErr
}

func NewService() IService {
	return &service{}
}

func (s *service) Create(dto *CreateEndpointDto, svc *corev1.Service, namespace string) *restErrors.RestErr {
	routes := make([]ingressroute.Route, 0)
	for _, v := range svc.Spec.Ports {
		if availableProtocol(v.Name) {
			routes = append(routes, ingressroute.Route{
				Match: fmt.Sprintf("Host(`endpoint.%s`) && Path(`/%s/%s`)", config.EnvironmentConf["DOMAIN_MATCH_BASE_URL"], svc.UID, v.Name),
				Services: []ingressroute.Service{
					{
						Name: svc.Name,
						Port: v.Port,
					},
				},
			})
		}
	}

	dtoIngressRoute := &ingressroute.IngressRoute{
		Name:      dto.Name,
		Namespace: namespace,
		Routes:    routes,
	}

	err := ingressRoutesService.Create(dtoIngressRoute)
	return err
}

var availableProtocol = func(protocol string) bool {
	switch protocol {
	case "ws":
		return false
	case "p2p":
		return false
	default:
		return true
	}
}
