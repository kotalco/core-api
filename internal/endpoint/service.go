package endpoint

import (
	"fmt"
	"github.com/kotalco/cloud-api/pkg/k8s/ingressroute"
	"github.com/kotalco/cloud-api/pkg/k8s/middleware"
	k8svc "github.com/kotalco/cloud-api/pkg/k8s/svc"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

var (
	ingressRoutesService = ingressroute.NewIngressRoutesService()
	k8MiddlewareService  = middleware.NewK8Middleware()
)

type service struct{}

type IService interface {
	//Create creates endpoint by:
	//-creating a middleware that get used by the ingressRoute to remove prefixes from the path before forwarding the request
	//-creating an ingressRoute (Traefik HTTP router) which uses the previous middleware
	//-return error if any
	Create(dto *CreateEndpointDto, svc *corev1.Service, namespace string) *restErrors.RestErr
	List(namespace string) ([]*EndpointDto, *restErrors.RestErr)
	Get(name string, namespace string) (*EndpointDto, *restErrors.RestErr)
	Delete(name string, namespace string) *restErrors.RestErr
}

func NewService() IService {
	return &service{}
}

func (s *service) Create(dto *CreateEndpointDto, svc *corev1.Service, namespace string) *restErrors.RestErr {
	ingressRoutePorts := make([]string, 0)
	middlewarePrefixes := make([]string, 0)

	for _, v := range svc.Spec.Ports {
		if k8svc.AvailableProtocol(v.Name) {
			ingressRoutePorts = append(ingressRoutePorts, v.Name)                                   //create ingressRoute ports
			middlewarePrefixes = append(middlewarePrefixes, fmt.Sprintf("/%s/%s", svc.UID, v.Name)) //create middleware prefixes
		}
	}

	//create the strip prefix-middleware
	middlewareName := fmt.Sprintf("%s-strip-prefix-%s", dto.Name, svc.UID)
	err := k8MiddlewareService.Create(&middleware.CreateMiddlewareDto{
		Name:          middlewareName,
		Namespace:     namespace,
		StripPrefixes: middlewarePrefixes,
	})
	if err != nil {
		return err
	}

	//create ingress-route
	err = ingressRoutesService.Create(&ingressroute.IngressRouteDto{
		Name:        dto.Name,
		Namespace:   namespace,
		ServiceName: svc.Name,
		ServiceID:   string(svc.UID),
		Ports:       ingressRoutePorts,
		Middlewares: []ingressroute.IngressRouteMiddlewareRefDto{{Name: middlewareName, Namespace: namespace}},
	})

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

func (s *service) Delete(name string, namespace string) *restErrors.RestErr {
	return ingressRoutesService.Delete(name, namespace)
}
