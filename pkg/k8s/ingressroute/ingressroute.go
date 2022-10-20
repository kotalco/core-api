package ingressroute

import (
	"context"
	"fmt"
	"github.com/kotalco/cloud-api/pkg/config"
	"github.com/kotalco/cloud-api/pkg/k8s"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)
import traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"

type ingressroute struct{}

// IIngressRoute has methods to work with ingressroutes resources.
type IIngressRoute interface {
	// Create takes the representation of a ingressRoute and creates it returns an error, if there is any.
	Create(dto *IngressRouteDto) *restErrors.RestErr
}

func NewIngressRoutesService() IIngressRoute {
	return &ingressroute{}
}

func (i *ingressroute) Create(dto *IngressRouteDto) *restErrors.RestErr {

	routes := make([]traefikv1alpha1.Route, 0)
	for k := 0; k < len(dto.Ports); k++ {
		routes = append(routes, traefikv1alpha1.Route{
			Match: fmt.Sprintf("Host(`endpoint.%s`) && Path(`/%s/%s`)", config.EnvironmentConf["DOMAIN_MATCH_BASE_URL"], dto.ServiceID, dto.Ports[k]),
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
			EntryPoints: []string{"web"},
			Routes:      routes,
		},
	}
	err := k8s.K8sClient.Create(context.Background(), route)
	if err != nil {
		go logger.Error(i.Create, err)
		return restErrors.NewInternalServerError(err.Error())
	}

	return nil
}
