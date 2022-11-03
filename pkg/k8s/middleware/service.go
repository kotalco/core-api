package middleware

import (
	"context"
	"fmt"
	"github.com/kotalco/cloud-api/pkg/k8s"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IK8Middleware interface {
	Create(dto *CreateMiddlewareDto) *restErrors.RestErr
}

type k8Middleware struct{}

func NewK8Middleware() IK8Middleware {
	return &k8Middleware{}
}

func (m *k8Middleware) Create(dto *CreateMiddlewareDto) *restErrors.RestErr {
	newMiddleware := &traefikv1alpha1.Middleware{
		ObjectMeta: metav1.ObjectMeta{
			Name:            dto.Name,
			Namespace:       dto.Namespace,
			OwnerReferences: dto.OwnersRef,
		},
		Spec: traefikv1alpha1.MiddlewareSpec{
			StripPrefix: &dynamic.StripPrefix{
				Prefixes: dto.StripPrefixes,
			},
		},
	}

	err := k8s.K8sClient.Create(context.Background(), newMiddleware)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return restErrors.NewConflictError(fmt.Sprintf("middleware %s already exist!", dto.Name))
		}
		go logger.Error(m.Create, err)
		return restErrors.NewInternalServerError(err.Error())
	}

	return nil
}
