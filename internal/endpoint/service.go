package endpoint

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/kotalco/cloud-api/pkg/k8s/ingressroute"
	"github.com/kotalco/cloud-api/pkg/k8s/middleware"
	"github.com/kotalco/cloud-api/pkg/k8s/secret"
	k8svc "github.com/kotalco/cloud-api/pkg/k8s/svc"
	"github.com/kotalco/cloud-api/pkg/security"
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
	Create(dto *CreateEndpointDto, svc *corev1.Service) restErrors.IRestErr
	List(namespace string) ([]*EndpointMetaDto, restErrors.IRestErr)
	Get(name string, namespace string) (*EndpointDto, restErrors.IRestErr)
	Delete(name string, namespace string) restErrors.IRestErr
}

func NewService() IService {
	return &service{}
}

func (s *service) Create(dto *CreateEndpointDto, svc *corev1.Service) restErrors.IRestErr {
	ingressRoutePorts := make([]ingressroute.IngressRoutePortDto, 0)
	middlewarePrefixes := make([]string, 0)
	stripePrefixMiddlewareName := fmt.Sprintf("%s-strip-prefix", dto.Name)
	basicAuthMiddlewareName := fmt.Sprintf("%s-basic-auth", dto.Name)

	for _, v := range svc.Spec.Ports {
		if k8svc.AvailableProtocol(v.Name) {
			ingressRoutePortDto := ingressroute.IngressRoutePortDto{
				ID:   uuid.NewString(),
				Name: v.Name,
			}
			ingressRoutePorts = append(ingressRoutePorts, ingressRoutePortDto)                          //create ingressRoute ports
			middlewarePrefixes = append(middlewarePrefixes, fmt.Sprintf("/%s", ingressRoutePortDto.ID)) //create middleware prefixes
		}
	}

	//create ingress-route
	component := strings.Split(svc.Labels["app.kubernetes.io/component"], "-")
	kind := component[len(component)-1]
	ingressRouteObject, err := ingressRoutesService.Create(&ingressroute.IngressRouteDto{
		Name:            dto.Name,
		Namespace:       svc.Namespace,
		ServiceName:     svc.Name,
		ServiceKind:     kind,
		ServiceProtocol: svc.Labels["kotal.io/protocol"],
		ServiceID:       string(svc.UID),
		Ports:           ingressRoutePorts,
		Middlewares: func() []ingressroute.IngressRouteMiddlewareRefDto {
			refs := make([]ingressroute.IngressRouteMiddlewareRefDto, 0)
			refs = append(refs, ingressroute.IngressRouteMiddlewareRefDto{
				Name:      stripePrefixMiddlewareName,
				Namespace: svc.Namespace,
			})
			if dto.UseBasicAuth {
				refs = append(refs, ingressroute.IngressRouteMiddlewareRefDto{
					Name:      basicAuthMiddlewareName,
					Namespace: svc.Namespace,
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
			Namespace:       svc.Namespace,
			OwnerReferences: []metav1.OwnerReference{ingressRouteOwnerRef},
		},
		MiddlewareSpec: v1alpha1.MiddlewareSpec{
			StripPrefix: &dynamic.StripPrefix{
				Prefixes: middlewarePrefixes,
			},
		},
	})
	if err != nil {
		dErr := ingressRoutesService.Delete(dto.Name, svc.Namespace)
		if dErr != nil {
			go logger.Error(s.Create, dErr)
		}
		return err
	}

	//create basic-auth middleware if exist
	if dto.UseBasicAuth {
		//create basic auth secret
		//since the endpoint name is unique, and we have 1 secret per endpoint
		//we can create secret name with the key secret+ endpointName
		secretName := fmt.Sprintf("%s-secret", dto.Name)
		err := secretService.Create(&secret.CreateSecretDto{
			ObjectMeta: metav1.ObjectMeta{
				Name:            secretName,
				Namespace:       svc.Namespace,
				OwnerReferences: []metav1.OwnerReference{ingressRouteOwnerRef},
			},
			Type: corev1.SecretTypeBasicAuth,
			StringData: map[string]string{
				"username": security.GenerateRandomString(8),
				"password": security.GenerateRandomString(8),
			},
		})
		if err != nil {
			dErr := ingressRoutesService.Delete(dto.Name, svc.Namespace)
			if dErr != nil {
				go logger.Error(s.Create, dErr)
			}
			return err
		}
		//create basic auth middleware
		err = k8MiddlewareService.Create(&middleware.CreateMiddlewareDto{
			ObjectMeta: metav1.ObjectMeta{
				Name:            basicAuthMiddlewareName,
				Namespace:       svc.Namespace,
				OwnerReferences: []metav1.OwnerReference{ingressRouteOwnerRef},
			},
			MiddlewareSpec: v1alpha1.MiddlewareSpec{
				BasicAuth: &v1alpha1.BasicAuth{
					Secret: secretName,
				},
			},
		})
		if err != nil {
			dErr := ingressRoutesService.Delete(dto.Name, svc.Namespace)
			if dErr != nil {
				go logger.Error(s.Create, dErr)
			}
			return err
		}
	}

	return nil
}

func (s *service) List(namespace string) ([]*EndpointMetaDto, restErrors.IRestErr) {
	records, err := ingressRoutesService.List(namespace)
	if err != nil {
		return nil, err
	}

	marshalledDto := make([]*EndpointMetaDto, 0)
	for _, item := range records.Items {
		marshalledDto = append(marshalledDto, new(EndpointMetaDto).Marshall(&item))
	}

	return marshalledDto, nil
}

func (s *service) Get(name string, namespace string) (*EndpointDto, restErrors.IRestErr) {
	record, err := ingressRoutesService.Get(name, namespace)
	if err != nil {
		return nil, err
	}

	//get secret
	secretName := fmt.Sprintf("%s-secret", record.Name)
	v1Secret, _ := secretService.Get(secretName, namespace)

	return new(EndpointDto).Marshall(record, v1Secret), nil
}

func (s *service) Delete(name string, namespace string) restErrors.IRestErr {
	return ingressRoutesService.Delete(name, namespace)
}
