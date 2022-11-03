package ingressroute

const (
	APIVersion = "traefik.containo.us/v1alpha1"
	Kind       = "IngressRoute"
)

// IngressRouteDto defines the desired state of the traefikv1alpha1.IngressRoute
// IngressRouteDto is the struct that matches the CRD implementation of a Traefik HTTP Router.
type IngressRouteDto struct {
	Name        string   // is the name of the ingress-route chosen by the user
	Namespace   string   // is the namespace the user wishes to create the ingress-route in
	ServiceName string   // is the name of the service created by kotal operator
	ServiceID   string   // id of the service created by kotal operator
	Ports       []string // the corresponding ports of the targeted service
	Middlewares []IngressRouteMiddlewareRefDto
}

type IngressRouteMiddlewareRefDto struct {
	Name      string
	Namespace string
}
