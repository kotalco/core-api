package endpoint

import (
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
	ports := make([]string, 0)
	for _, v := range svc.Spec.Ports {
		if availableProtocol(v.Name) {
			ports = append(ports, v.Name)
		}
	}

	ingresRouteDto := &ingressroute.IngressRouteDto{
		Name:        dto.Name,
		Namespace:   namespace,
		ServiceName: svc.Name,
		ServiceID:   string(svc.UID),
		Ports:       ports,
	}

	err := ingressRoutesService.Create(ingresRouteDto)
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
