package endpoint

import (
	"encoding/json"
	"fmt"
	"github.com/kotalco/core-api/config"
	ingressroute2 "github.com/kotalco/core-api/k8s/ingressroute"
	middleware2 "github.com/kotalco/core-api/k8s/middleware"
	secret2 "github.com/kotalco/core-api/k8s/secret"
	k8svc "github.com/kotalco/core-api/k8s/svc"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"net/http"
	"strconv"
	"strings"

	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/logger"
	"github.com/kotalco/core-api/pkg/security"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// The crossover middleware is a plugin used to store statistics for each endpoint, utilizing the endpointActivity package
const (
	crossoverActivityMiddlewareName = "crossover-activity"
	crossoverCacheMiddlewareName    = "crossover-cache"
	crossoverMiddlewareNamespace    = "kotal"
)

var (
	ingressRoutesService = ingressroute2.NewIngressRoutesService()
	k8MiddlewareService  = middleware2.NewK8Middleware()
	secretService        = secret2.NewService()
)

type service struct{}

type IService interface {
	//Create creates endpoint by:
	//-creating a middleware that get used by the ingressRoute to remove prefixes from the path before forwarding the request
	//-creating an ingressRoute (Traefik HTTP router) which uses the previous middleware
	//-return error if any
	Create(dto *CreateEndpointDto, svc *corev1.Service) restErrors.IRestErr
	List(ns string, labels map[string]string) (*v1alpha1.IngressRouteList, restErrors.IRestErr)
	Get(name string, namespace string) (*v1alpha1.IngressRoute, restErrors.IRestErr)
	Delete(name string, namespace string) restErrors.IRestErr
	Count(ns string, labels map[string]string) (int, restErrors.IRestErr)
}

func NewService() IService {
	return &service{}
}

func (s *service) Create(dto *CreateEndpointDto, svc *corev1.Service) restErrors.IRestErr {
	ingressRoutePorts := make([]ingressroute2.IngressRoutePortDto, 0)
	middlewarePrefixes := make([]string, 0)
	stripePrefixMiddlewareName := fmt.Sprintf("%s-strip-prefix", dto.Name)
	basicAuthMiddlewareName := fmt.Sprintf("%s-basic-auth", dto.Name)
	endpointPortIdLength, intErr := strconv.Atoi(config.Environment.EndpointPortIdLength)
	if intErr != nil {
		return restErrors.NewInternalServerError(intErr.Error())
	}
	for _, v := range svc.Spec.Ports {
		if k8svc.AvailableProtocol(v.Name) {
			ingressRoutePortDto := ingressroute2.IngressRoutePortDto{
				ID:   fmt.Sprintf("%s%s", strings.ToLower(security.GenerateRandomString(endpointPortIdLength)), strings.Replace(dto.UserId, "-", "", -1)),
				Name: v.Name,
			}
			ingressRoutePorts = append(ingressRoutePorts, ingressRoutePortDto)                          //create ingressRoute ports
			middlewarePrefixes = append(middlewarePrefixes, fmt.Sprintf("/%s", ingressRoutePortDto.ID)) //create middleware prefixes
		}
	}

	//create ingress-route
	ingressRouteObject, err := ingressRoutesService.Create(&ingressroute2.IngressRouteDto{
		Name:        dto.Name,
		Namespace:   svc.Namespace,
		ServiceName: svc.Name,
		ServiceID:   string(svc.UID),
		Ports:       ingressRoutePorts,
		Middlewares: func() []ingressroute2.IngressRouteMiddlewareRefDto {
			refs := make([]ingressroute2.IngressRouteMiddlewareRefDto, 0)
			//append crossover  middleware
			refs = append(refs, ingressroute2.IngressRouteMiddlewareRefDto{
				Name:      crossoverActivityMiddlewareName,
				Namespace: crossoverMiddlewareNamespace,
			})
			//append crossover  middleware
			refs = append(refs, ingressroute2.IngressRouteMiddlewareRefDto{
				Name:      crossoverCacheMiddlewareName,
				Namespace: crossoverMiddlewareNamespace,
			})
			//append stripePrefix middleware
			refs = append(refs, ingressroute2.IngressRouteMiddlewareRefDto{
				Name:      stripePrefixMiddlewareName,
				Namespace: svc.Namespace,
			})
			//append basicAuth middleware
			if dto.UseBasicAuth {
				refs = append(refs, ingressroute2.IngressRouteMiddlewareRefDto{
					Name:      basicAuthMiddlewareName,
					Namespace: svc.Namespace,
				})
			}
			return refs
		}(),
		OwnersRef: svc.OwnerReferences,
		Labels: func() map[string]string {
			if dto.Labels == nil {
				dto.Labels = map[string]string{}
			}
			component := strings.Split(svc.Labels["app.kubernetes.io/component"], "-")
			dto.Labels["app.kubernetes.io/created-by"] = "kotal-api"
			dto.Labels["kotal.io/protocol"] = svc.Labels["kotal.io/protocol"]
			dto.Labels["kotal.io/network"] = svc.Labels["kotal.io/network"]
			dto.Labels["kotal.io/user-id"] = dto.UserId
			dto.Labels["kotal.io/kind"] = component[len(component)-1]
			return dto.Labels
		}(),
	})
	if err != nil {
		return err
	}

	ingressRouteOwnerRef := metav1.OwnerReference{
		APIVersion: ingressroute2.APIVersion,
		Kind:       ingressroute2.Kind,
		Name:       ingressRouteObject.Name,
		UID:        ingressRouteObject.UID,
	}

	//create the strip prefix-middleware
	err = k8MiddlewareService.Create(&middleware2.CreateMiddlewareDto{
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
			return dErr
		}
		return err
	}

	//create basic-auth middleware if exist
	if dto.UseBasicAuth {
		//create basic auth secret
		//since the endpoint name is unique, and we have 1 secret per endpoint
		//we can create secret name with the key secret+ endpointName
		secretName := fmt.Sprintf("%s-secret", dto.Name)
		err := secretService.Create(&secret2.CreateSecretDto{
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
				return dErr
			}
			return err
		}
		//create basic auth middleware
		err = k8MiddlewareService.Create(&middleware2.CreateMiddlewareDto{
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

	//create crossover activity middleware if it doesn't exist
	_, err = k8MiddlewareService.Get(crossoverActivityMiddlewareName, crossoverMiddlewareNamespace)
	if err != nil {
		if err.StatusCode() == http.StatusNotFound {
			jsonBytes, intErr := json.Marshal(map[string]interface{}{
				"APIKey":        config.Environment.CrossOverAPIKey,
				"Pattern":       config.Environment.CrossOverPattern,
				"RemoteAddress": config.Environment.CrossOverRemoteAddress,
				"BufferSize":    config.Environment.CrossOverActivityBufferSize,
				"BatchSize":     config.Environment.CrossOverActivityBatchSize,
				"FlushInterval": config.Environment.CrossOverActivityFlushInterval,
			})
			if intErr != nil {
				go logger.Error(s.Create, intErr)
				return restErrors.NewInternalServerError("something went wrong")
			}

			err = k8MiddlewareService.Create(&middleware2.CreateMiddlewareDto{
				ObjectMeta: metav1.ObjectMeta{
					Name:      crossoverActivityMiddlewareName,
					Namespace: crossoverMiddlewareNamespace,
				},
				MiddlewareSpec: v1alpha1.MiddlewareSpec{
					Plugin: map[string]v1.JSON{
						crossoverActivityMiddlewareName: {
							Raw: jsonBytes,
						},
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
		} else {
			dErr := ingressRoutesService.Delete(dto.Name, svc.Namespace)
			if dErr != nil {
				go logger.Error(s.Create, dErr)
			}
			return err
		}
	}

	//create crossover cache middleware if it doesn't exist
	_, err = k8MiddlewareService.Get(crossoverCacheMiddlewareName, crossoverMiddlewareNamespace)
	if err != nil {
		if err.StatusCode() == http.StatusNotFound {
			jsonBytes, intErr := json.Marshal(map[string]interface{}{
				"RedisAddress":  config.Environment.CrossOverRedisAddress,
				"RedisAuth":     config.Environment.CrossOverRedisAuth,
				"RedisPoolSize": config.Environment.CrossOverRedisPoolSize,
				"CacheExpiry":   config.Environment.CrossOverRedisCacheExpiry,
			})
			if intErr != nil {
				go logger.Error(s.Create, intErr)
				return restErrors.NewInternalServerError("something went wrong")
			}

			err = k8MiddlewareService.Create(&middleware2.CreateMiddlewareDto{
				ObjectMeta: metav1.ObjectMeta{
					Name:      crossoverCacheMiddlewareName,
					Namespace: crossoverMiddlewareNamespace,
				},
				MiddlewareSpec: v1alpha1.MiddlewareSpec{
					Plugin: map[string]v1.JSON{
						crossoverCacheMiddlewareName: {
							Raw: jsonBytes,
						},
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
		} else {
			dErr := ingressRoutesService.Delete(dto.Name, svc.Namespace)
			if dErr != nil {
				go logger.Error(s.Create, dErr)
			}
			return err
		}
	}
	return nil
}

func (s *service) List(ns string, labels map[string]string) (*v1alpha1.IngressRouteList, restErrors.IRestErr) {
	return ingressRoutesService.List(ns, labels)
}

func (s *service) Get(name string, namespace string) (*v1alpha1.IngressRoute, restErrors.IRestErr) {
	return ingressRoutesService.Get(name, namespace)
}

func (s *service) Delete(name string, namespace string) restErrors.IRestErr {
	return ingressRoutesService.Delete(name, namespace)
}

func (s *service) Count(ns string, labels map[string]string) (count int, err restErrors.IRestErr) {
	records, err := ingressRoutesService.List(ns, labels)
	if err != nil {
		return 0, err
	}
	return len(records.Items), err
}
