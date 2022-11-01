package middleware

import (
	"context"
	"github.com/kotalco/cloud-api/pkg/k8s"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const StripPrefixRegexMiddlewareName = "stripprefixregex"

func CreateStripPrefixRegexMiddleware() *restErrors.RestErr {
	newMiddleware := &traefikv1alpha1.Middleware{
		ObjectMeta: metav1.ObjectMeta{
			Name:      StripPrefixRegexMiddlewareName,
			Namespace: "default",
		},
		Spec: traefikv1alpha1.MiddlewareSpec{
			StripPrefixRegex: &dynamic.StripPrefixRegex{Regex: []string{"/([A-Za-z0-9]+(-[A-Za-z0-9]+)+)/[A-Za-z0-9]+"}},
		},
	}

	err := k8s.K8sClient.Create(context.Background(), newMiddleware)
	if err != nil {
		if errors.IsConflict(err) {
			go logger.Error("CreateStripPrefixRegexMiddleware", err)
			return restErrors.NewConflictError(err.Error())
		}
		return restErrors.NewInternalServerError(err.Error())
	}

	return nil
}
