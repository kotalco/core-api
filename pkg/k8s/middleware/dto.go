package middleware

import (
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CreateMiddlewareDto struct {
	metav1.ObjectMeta
	dynamic.StripPrefix
	OwnersRef []metav1.OwnerReference
}
