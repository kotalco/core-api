package k8s

import (
	"context"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/k8s"
	"github.com/kotalco/community-api/pkg/logger"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)
import traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"

type endpoint struct{}
type IEndpoint interface {
	Create(name string, namespace string, serviceName string, port int) *restErrors.RestErr
	List(namespace string) (*traefikv1alpha1.IngressRouteList, *restErrors.RestErr)
	Get(name string, namespace string) (*v1.Ingress, *restErrors.RestErr)
}

func NewEndpointService() IEndpoint {
	return &endpoint{}
}

func (service *endpoint) Create(name string, namespace string, serviceName string, port int) *restErrors.RestErr {
	route := &traefikv1alpha1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: traefikv1alpha1.IngressRouteSpec{
			EntryPoints: []string{"web"},
			Routes: []traefikv1alpha1.Route{{
				Match: "Host(`nginx.example.com`)",
				Kind:  "Rule",
				Services: []traefikv1alpha1.Service{{
					LoadBalancerSpec: traefikv1alpha1.LoadBalancerSpec{
						Name:   serviceName,
						Sticky: nil,
						Port:   intstr.IntOrString{IntVal: 80},
					},
				}},
			}},
		},
	}
	err := k8sClient.Create(context.Background(), route)
	if err != nil {
		go logger.Error(service.Create, err)
		return restErrors.NewInternalServerError(err.Error())
	}

	return nil
}

func (service *endpoint) List(namespace string) (*traefikv1alpha1.IngressRouteList, *restErrors.RestErr) {
	nodes := &traefikv1alpha1.IngressRouteList{}
	err := k8sClient.List(context.Background(), nodes, client.InNamespace(namespace))
	if err != nil {
		go logger.Error(service.List, err)
		return nil, restErrors.NewInternalServerError("failed to get all nodes")
	}

	return nodes, nil
}

func (service *endpoint) Get(name string, namespace string) (*v1.Ingress, *restErrors.RestErr) {
	record, err := k8s.Clientset().NetworkingV1().Ingresses(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		go logger.Error(service.Get, err)
		return nil, restErrors.NewInternalServerError(err.Error())
	}
	return record, nil
}
