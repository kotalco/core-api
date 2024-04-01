package ingressroute

import (
	"context"
	"fmt"
	"github.com/kotalco/core-api/config"
	"github.com/kotalco/core-api/k8s"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/logger"
	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

var k8sClient = k8s.NewClientService()

type IngressRoute interface {
	Get() (*traefikv1alpha1.IngressRoute, restErrors.IRestErr)
	SetCertResolver(resolverName string) restErrors.IRestErr
	SetTLSSecret(secretName string) restErrors.IRestErr
}

type ingressRoute struct {
}

func NewService() IngressRoute {

	return &ingressRoute{}
}

func (s *ingressRoute) Get() (*traefikv1alpha1.IngressRoute, restErrors.IRestErr) {
	var record traefikv1alpha1.IngressRoute
	key := types.NamespacedName{
		Namespace: config.Environment.KotalNamespace,
		Name:      config.Environment.KotalIngressRouteName,
	}

	err := k8sClient.Get(context.Background(), key, &record)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, restErrors.NewNotFoundError(fmt.Sprintf("can't find endpoint %s", record.Name))
		}
		go logger.Error(s.Get, err)
		return nil, restErrors.NewInternalServerError("something went wrong")
	}
	return &record, nil
}

func (s *ingressRoute) SetCertResolver(resolverName string) restErrors.IRestErr {
	ingressRoute, restErr := s.Get()
	if restErr != nil {
		return restErr
	}
	ingressRoute.Spec.TLS = &traefikv1alpha1.TLS{
		CertResolver: resolverName,
	}
	err := k8sClient.Update(context.Background(), ingressRoute)
	if err != nil {
		go logger.Warn(s.SetCertResolver, err)
		return restErrors.NewInternalServerError(err.Error())
	}
	return nil
}

func (s *ingressRoute) SetTLSSecret(secretName string) restErrors.IRestErr {
	ingressRoute, restErr := s.Get()
	if restErr != nil {
		return restErr
	}
	ingressRoute.Spec.TLS = &traefikv1alpha1.TLS{
		SecretName: secretName,
	}
	err := k8sClient.Update(context.Background(), ingressRoute)
	if err != nil {
		go logger.Warn(s.SetTLSSecret, err)
		return restErrors.NewInternalServerError(err.Error())
	}
	return nil
}
