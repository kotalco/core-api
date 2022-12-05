package ingressroute

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	APIVersion = "traefik.containo.us/v1alpha1"
	Kind       = "IngressRoute"
)

// IngressRouteDto defines the desired state of the traefikv1alpha1.IngressRoute
// IngressRouteDto is the struct that matches the CRD implementation of a Traefik HTTP Router.
type IngressRouteDto struct {
	Name            string                // is the name of the ingress-route chosen by the user
	Namespace       string                // is the namespace the user wishes to create the ingress-route in
	ServiceName     string                // is the name of the service created by kotal operator
	ServiceProtocol string                // is the protocol of which responsible fo creating this service
	ServiceID       string                // id of the service created by kotal operator
	Ports           []IngressRoutePortDto // the corresponding ports of the targeted service
	Middlewares     []IngressRouteMiddlewareRefDto
	OwnersRef       []metav1.OwnerReference
}

type IngressRouteMiddlewareRefDto struct {
	Name      string
	Namespace string
}

type IngressRoutePortDto struct {
	ID   string
	Name string
}
