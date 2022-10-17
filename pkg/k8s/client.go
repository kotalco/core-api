package k8s

import (
	"github.com/kotalco/community-api/pkg/k8s"
	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
)

var k8sClient = func() k8s.K8sClientServiceInterface {

	traefikv1alpha1.AddToScheme(k8s.RunTimeScheme)
	return k8s.NewClientService()
}()
