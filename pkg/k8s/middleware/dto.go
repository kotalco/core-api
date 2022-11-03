package middleware

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type CreateMiddlewareDto struct {
	Name          string
	Namespace     string
	StripPrefixes []string
	OwnersRef     []metav1.OwnerReference
}
