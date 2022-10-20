package endpoint

import (
	"github.com/kotalco/cloud-api/pkg/k8s/ingressroute"
	k8svc "github.com/kotalco/cloud-api/pkg/k8s/svc"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

var (
	ingressRoutesService = ingressroute.NewIngressRoutesService()
)

type service struct{}

type IService interface {
	Create(dto *CreateEndpointDto, svc *corev1.Service, namespace string) *restErrors.RestErr
	List(namespace string) ([]*EndpointDto, *restErrors.RestErr)
	Get(name string, namespace string) (*EndpointDto, *restErrors.RestErr)
	Update(name string, namespace string, newName string) *restErrors.RestErr
	Delete(name string, namespace string) *restErrors.RestErr
}

func NewService() IService {
	return &service{}
}

func (s *service) Create(dto *CreateEndpointDto, svc *corev1.Service, namespace string) *restErrors.RestErr {
	ports := make([]string, 0)
	for _, v := range svc.Spec.Ports {
		if k8svc.AvailableProtocol(v.Name) {
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

func (s *service) List(namespace string) ([]*EndpointDto, *restErrors.RestErr) {
	records, err := ingressRoutesService.List(namespace)
	if err != nil {
		return nil, err
	}

	marshalledDto := make([]*EndpointDto, 0)
	for _, item := range records.Items {
		marshalledDto = append(marshalledDto, new(EndpointDto).Marshall(&item))
	}

	return marshalledDto, nil
}

func (s *service) Get(name string, namespace string) (*EndpointDto, *restErrors.RestErr) {
	record, err := ingressRoutesService.Get(name, namespace)
	if err != nil {
		return nil, err
	}

	return new(EndpointDto).Marshall(record), nil
}

func (s *service) Update(name string, namespace string, newName string) *restErrors.RestErr {
	return ingressRoutesService.Update(name, namespace, newName)
}
func (s *service) Delete(name string, namespace string) *restErrors.RestErr {
	return ingressRoutesService.Delete(name, namespace)
}
