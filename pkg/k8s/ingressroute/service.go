package ingressroute

import (
	"context"
	"fmt"
	"github.com/kotalco/cloud-api/core/setting"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kotalco/cloud-api/pkg/k8s"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ingressroute struct{}

// IIngressRoute has methods to work with ingressroutes resources.
type IIngressRoute interface {
	// Create takes the representation of a ingressRoute and creates it returns ingress-route object or error if any
	Create(dto *IngressRouteDto) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr)
	//List takes label and field selectors, and returns the list of Middlewares that match those selectors.
	List(ns string, labels map[string]string) (*traefikv1alpha1.IngressRouteList, restErrors.IRestErr)
	// Get takes name and namespace of the ingressRoute, and returns the corresponding ingressRoute object, and an error if there is any.
	Get(name string, namespace string) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr)
	// Delete takes name  and namespace of the ingressRoute, check if it exists and delete it if found. Returns an error if one occurs.
	Delete(name string, namespace string) restErrors.IRestErr
	//Update takes IngressRoute and updates it, return an error if any
	Update(record *traefikv1alpha1.IngressRoute) restErrors.IRestErr
}

func NewIngressRoutesService() IIngressRoute {
	return &ingressroute{}
}

func (i *ingressroute) Create(dto *IngressRouteDto) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr) {
	domainBaseUrl, restErr := setting.GetDomainBaseUrl()
	if restErr != nil {
		return nil, restErr
	}
	routes := make([]traefikv1alpha1.Route, 0)
	for k := 0; k < len(dto.Ports); k++ {
		routes = append(routes, traefikv1alpha1.Route{
			Match: fmt.Sprintf("Host(`endpoints.%s`) && PathPrefix(`/%s`)", domainBaseUrl, dto.Ports[k].ID),
			Kind:  "Rule",
			Middlewares: func() []traefikv1alpha1.MiddlewareRef {
				middlewares := make([]traefikv1alpha1.MiddlewareRef, 0)
				for _, v := range dto.Middlewares {
					middlewares = append(middlewares, traefikv1alpha1.MiddlewareRef{Name: v.Name, Namespace: v.Namespace})
				}
				middlewares = append(middlewares, traefikv1alpha1.MiddlewareRef{Name: "my-crossover", Namespace: "default"})
				return middlewares
			}(),
			Services: []traefikv1alpha1.Service{
				{
					LoadBalancerSpec: traefikv1alpha1.LoadBalancerSpec{
						Name:      dto.ServiceName,
						Namespace: dto.Namespace, // find the service in the namespace of the ingressroute
						Port:      intstr.IntOrString{Type: intstr.String, StrVal: dto.Ports[k].Name},
					},
				},
			},
		})
	}

	record := &traefikv1alpha1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:            dto.Name,
			Namespace:       dto.Namespace,
			OwnerReferences: dto.OwnersRef,
			Labels:          map[string]string{"app.kubernetes.io/created-by": "kotal-api", "kotal.io/protocol": dto.ServiceProtocol, "kotal.io/network": dto.ServiceNetwork, "kotal.io/kind": dto.ServiceKind, "kotal.io/user-id": dto.UserId},
		},
		Spec: traefikv1alpha1.IngressRouteSpec{
			EntryPoints: []string{"websecure"},
			Routes:      routes,
			TLS: &traefikv1alpha1.TLS{
				CertResolver: "myresolver",
			},
		},
	}
	intErr := k8s.K8sClient.Create(context.Background(), record)
	if intErr != nil {
		if errors.IsAlreadyExists(intErr) {
			return nil, restErrors.NewConflictError(fmt.Sprintf("endpoint %s already exist!", dto.Name))
		}
		go logger.Error(i.Create, intErr)
		return nil, restErrors.NewInternalServerError(intErr.Error())
	}

	return record, nil
}

func (i *ingressroute) List(ns string, labels map[string]string) (*traefikv1alpha1.IngressRouteList, restErrors.IRestErr) {
	records := &traefikv1alpha1.IngressRouteList{}
	var matchingLabels = client.MatchingLabels{}
	matchingLabels = labels
	err := k8s.K8sClient.List(context.Background(), records, &client.ListOptions{Namespace: ns}, &matchingLabels)
	if err != nil {
		go logger.Error(i.List, err)
	}
	return records, nil
}

func (i *ingressroute) Get(name string, namespace string) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr) {
	var record traefikv1alpha1.IngressRoute
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}

	err := k8s.K8sClient.Get(context.Background(), key, &record)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, restErrors.NewNotFoundError(fmt.Sprintf("can't find endpoint %s", name))
		}
		go logger.Error(i.Get, err)
		return nil, restErrors.NewInternalServerError("something went wrong")
	}
	return &record, nil
}

func (i *ingressroute) Delete(name string, namespace string) restErrors.IRestErr {
	record, err := i.Get(name, namespace)
	if err != nil {
		return err
	}

	intErr := k8s.K8sClient.Delete(context.Background(), record)
	if intErr != nil {
		go logger.Error(i.Delete, intErr)
		return restErrors.NewInternalServerError("something went wrong")
	}

	return nil
}

func (i *ingressroute) Update(record *traefikv1alpha1.IngressRoute) restErrors.IRestErr {
	intErr := k8s.K8sClient.Update(context.Background(), record)
	if intErr != nil {
		go logger.Error(i.Update, intErr)
		return restErrors.NewInternalServerError("can't update ingressRoute")
	}
	return nil
}
