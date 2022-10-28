package ingressroute

import (
	"context"
	"fmt"
	"github.com/kotalco/cloud-api/pkg/config"
	"github.com/kotalco/cloud-api/pkg/k8s"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ingressroute struct{}

// IIngressRoute has methods to work with ingressroutes resources.
type IIngressRoute interface {
	// Create takes the representation of a ingressRoute and creates it returns an error, if there is any.
	Create(dto *IngressRouteDto) *restErrors.RestErr
	//List takes label and field selectors, and returns the list of Middlewares that match those selectors.
	List(namesapce string) (*traefikv1alpha1.IngressRouteList, *restErrors.RestErr)
	// Get takes name and namespace of the ingressRoute, and returns the corresponding ingressRoute object, and an error if there is any.
	Get(name string, namespace string) (*traefikv1alpha1.IngressRoute, *restErrors.RestErr)
	// Delete takes name  and namespace of the ingressRoute, check if it exists and delete it if found. Returns an error if one occurs.
	Delete(name string, namespace string) *restErrors.RestErr
}

func NewIngressRoutesService() IIngressRoute {
	return &ingressroute{}
}

func (i *ingressroute) Create(dto *IngressRouteDto) *restErrors.RestErr {
	routes := make([]traefikv1alpha1.Route, 0)
	for k := 0; k < len(dto.Ports); k++ {
		routes = append(routes, traefikv1alpha1.Route{
			Match: fmt.Sprintf("Host(`endpoints.%s`) && Path(`/%s/%s`)", config.EnvironmentConf["DOMAIN_MATCH_BASE_URL"], dto.ServiceID, dto.Ports[k]),
			Kind:  "Rule",
			Services: []traefikv1alpha1.Service{
				{
					LoadBalancerSpec: traefikv1alpha1.LoadBalancerSpec{
						Name: dto.ServiceName,
						Port: intstr.IntOrString{Type: intstr.String, StrVal: dto.Ports[k]},
					},
				},
			},
		})
	}

	route := &traefikv1alpha1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dto.Name,
			Namespace: dto.Namespace,
		},
		Spec: traefikv1alpha1.IngressRouteSpec{
			EntryPoints: []string{"websecure"},
			Routes:      routes,
			TLS: &traefikv1alpha1.TLS{
				CertResolver: "myresolver",
			},
		},
	}
	err := k8s.K8sClient.Create(context.Background(), route)
	if err != nil {
		go logger.Error(i.Create, err)
		return restErrors.NewInternalServerError(err.Error())
	}

	return nil
}

func (i *ingressroute) List(namespace string) (*traefikv1alpha1.IngressRouteList, *restErrors.RestErr) {
	records := &traefikv1alpha1.IngressRouteList{}
	err := k8s.K8sClient.List(context.Background(), records, &client.ListOptions{Namespace: namespace})
	if err != nil {
		go logger.Error(i.List, err)
	}
	return records, nil
}

func (i *ingressroute) Get(name string, namespace string) (*traefikv1alpha1.IngressRoute, *restErrors.RestErr) {
	record := &traefikv1alpha1.IngressRoute{}
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	err := k8s.K8sClient.Get(context.Background(), key, record)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, restErrors.NewNotFoundError(fmt.Sprintf("can't find endpoint %s", name))
		}
		go logger.Error(i.Get, err)
		return nil, restErrors.NewInternalServerError("something went wrong")
	}
	return record, nil
}

func (i *ingressroute) Delete(name string, namespace string) *restErrors.RestErr {
	record, err := i.Get(name, namespace)
	if err != nil {
		return err
	}

	intErr := k8s.K8sClient.Delete(context.Background(), record)
	if intErr != nil {
		go logger.Error(i.Delete, err)
		return restErrors.NewInternalServerError("something went wrong")
	}

	return nil
}
