package endpoint

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/kotalco/cloud-api/pkg/k8s/ingressroute"
	"github.com/kotalco/cloud-api/pkg/k8s/middleware"
	"github.com/kotalco/cloud-api/pkg/k8s/secret"
	k8svc "github.com/kotalco/cloud-api/pkg/k8s/svc"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ingressRoutesService = ingressroute.NewIngressRoutesService()
	k8MiddlewareService  = middleware.NewK8Middleware()
	secretService        = secret.NewService()
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
	stripePrefixMiddlewareName := fmt.Sprintf("%s-strip-prefix-%s", dto.Name, uuid.NewString())
	basicAuthMiddlewareName := fmt.Sprintf("%s-basic-auth-%s", dto.Name, uuid.NewString())

	for _, v := range svc.Spec.Ports {
		if k8svc.AvailableProtocol(v.Name) {
			ingressRoutePorts = append(ingressRoutePorts, v.Name)                                   //create ingressRoute ports
			middlewarePrefixes = append(middlewarePrefixes, fmt.Sprintf("/%s/%s", svc.UID, v.Name)) //create middleware prefixes
		}
	}

	//create ingress-route
	ingressRouteObject, err := ingressRoutesService.Create(&ingressroute.IngressRouteDto{
		Name:        dto.Name,
		Namespace:   namespace,
		ServiceName: svc.Name,
		ServiceID:   string(svc.UID),
		Ports:       ingressRoutePorts,
		Middlewares: func() []ingressroute.IngressRouteMiddlewareRefDto {
			refs := make([]ingressroute.IngressRouteMiddlewareRefDto, 0)
			refs = append(refs, ingressroute.IngressRouteMiddlewareRefDto{
				Name:      stripePrefixMiddlewareName,
				Namespace: namespace,
			})
			if dto.BasicAuth != nil {
				refs = append(refs, ingressroute.IngressRouteMiddlewareRefDto{
					Name:      basicAuthMiddlewareName,
					Namespace: namespace,
				})
			}
			return refs
		}(),
		OwnersRef: svc.OwnerReferences,
	})
	if err != nil {
		return err
	}

	ingressRouteOwnerRef := metav1.OwnerReference{
		APIVersion: ingressroute.APIVersion,
		Kind:       ingressroute.Kind,
		Name:       ingressRouteObject.Name,
		UID:        ingressRouteObject.UID,
	}

	//create the strip prefix-middleware
	err = k8MiddlewareService.Create(&middleware.CreateMiddlewareDto{
		ObjectMeta: metav1.ObjectMeta{
			Name:            stripePrefixMiddlewareName,
			Namespace:       namespace,
			OwnerReferences: []metav1.OwnerReference{ingressRouteOwnerRef},
		},
		MiddlewareSpec: v1alpha1.MiddlewareSpec{
			StripPrefix: &dynamic.StripPrefix{
				Prefixes: middlewarePrefixes,
			},
		},
	})
	if err != nil {
		dErr := ingressRoutesService.Delete(dto.Name, namespace)
		if dErr != nil {
			go logger.Error(s.Create, dErr)
		}
		return err
	}

	//create basic-auth middleware if exist
	if dto.BasicAuth != nil {
		//create basic auth secret
		secretName := fmt.Sprintf("%s-secret-%s", dto.Name, svc.UID)
		err := secretService.Create(&secret.CreateSecretDto{
			ObjectMeta: metav1.ObjectMeta{
				Name:            secretName,
				Namespace:       namespace,
				OwnerReferences: []metav1.OwnerReference{ingressRouteOwnerRef},
			},
			Type: corev1.SecretTypeBasicAuth,
			StringData: map[string]string{
				"username": dto.BasicAuth.Username,
				"password": dto.BasicAuth.Password,
			},
		})
		if err != nil {
			dErr := ingressRoutesService.Delete(dto.Name, namespace)
			if dErr != nil {
				go logger.Error(s.Create, dErr)
			}
			return err
		}
		//create basic auth middleware
		err = k8MiddlewareService.Create(&middleware.CreateMiddlewareDto{
			ObjectMeta: metav1.ObjectMeta{
				Name:            basicAuthMiddlewareName,
				Namespace:       namespace,
				OwnerReferences: []metav1.OwnerReference{ingressRouteOwnerRef},
			},
			MiddlewareSpec: v1alpha1.MiddlewareSpec{
				BasicAuth: &v1alpha1.BasicAuth{
					Secret: secretName,
				},
			},
		})
		if err != nil {
			dErr := ingressRoutesService.Delete(dto.Name, namespace)
			if dErr != nil {
				go logger.Error(s.Create, dErr)
			}
			return err
		}
	}

	return nil
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
