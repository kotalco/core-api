package k8s

import (
	"github.com/kotalco/community-api/pkg/k8s"
	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
)

var K8sClient = func() k8s.K8sClientServiceInterface {
	return k8s.NewClientService()
}()

func Config() {
	//AddToScheme used to add the additional types used be the cloud-api to the community-api runtime schema
	//should be called before  attempt to use the  community-api k8s.NewClientService coz k8s.NewClientService is a singleton
	traefikv1alpha1.AddToScheme(k8s.RunTimeScheme)
}
