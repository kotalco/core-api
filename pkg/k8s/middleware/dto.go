package middleware

import (
	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CreateMiddlewareDto struct {
	metav1.ObjectMeta
	traefikv1alpha1.MiddlewareSpec
}
