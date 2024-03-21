package kotal_ingressroute

import (
	"context"
	"fmt"
	"github.com/kotalco/core-api/k8s"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/logger"
	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

var k8sClient = k8s.NewClientService()

type KotalIR interface {
	Get() (*traefikv1alpha1.IngressRoute, restErrors.IRestErr)
	SetCertResolver(ingressRoute *traefikv1alpha1.IngressRoute, key string) restErrors.IRestErr
}

type kotalIR struct {
}

func NewService() KotalIR {

	return &kotalIR{}
}

func (s *kotalIR) Get() (*traefikv1alpha1.IngressRoute, restErrors.IRestErr) {
	var record traefikv1alpha1.IngressRoute
	key := types.NamespacedName{
		Namespace: "kotal",
		Name:      "kotal-stack",
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

func (s *kotalIR) SetCertResolver(ingressRoute *traefikv1alpha1.IngressRoute, key string) restErrors.IRestErr {
	ingressRoute.Spec.TLS = &traefikv1alpha1.TLS{
		CertResolver: key,
	}
	err := k8sClient.Update(context.Background(), ingressRoute)
	if err != nil {
		go logger.Warn("CONFIGURE_TLS", err)
		return restErrors.NewInternalServerError(err.Error())
	}
	return nil
}

func (s *kotalIR) SetTLSSecret(ingressRoute *traefikv1alpha1.IngressRoute, secretName string) restErrors.IRestErr {
	ingressRoute.Spec.TLS = &traefikv1alpha1.TLS{
		SecretName: secretName,
	}
	err := k8sClient.Update(context.Background(), ingressRoute)
	if err != nil {
		go logger.Warn("CONFIGURE_TLS", err)
		return restErrors.NewInternalServerError(err.Error())
	}
	return nil
}
