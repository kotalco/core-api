package ingressroute

// IngressRoute defines the desired state of the traefikv1alpha1.IngressRoute
// IngressRoute is the struct that matches the CRD implementation of a Traefik HTTP Router.
type IngressRoute struct {
	Name      string // is the name of the ingress-route chosen by the user
	Namespace string // si the namespace the user wishes to create the ingress-route in
	Routes    []Route
}

// Routes holds the HTTP route configuration for each rotue
type Route struct {
	Match    string //match is the user baseDomain+path eg. Host(`endpoint.${domain}`) && Path(`/${service.uid}/rpc`)
	Services []Service
}

// Service defines an upstream HTTP service to proxy traffic to.
// It should contian the name of service created by kotal operator and it's corresponding port
//each service might have multiple ports
type Service struct {
	Name string
	Port int32
}
