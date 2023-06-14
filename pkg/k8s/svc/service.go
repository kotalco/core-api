package svc

import (
	"context"
	"fmt"
	"github.com/kotalco/cloud-api/pkg/k8s"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ISVC interface {
	List(namespace string) (*corev1.ServiceList, restErrors.IRestErr)
	Get(name string, namespace string) (*corev1.Service, restErrors.IRestErr)
	Create(obj *corev1.Service) restErrors.IRestErr
}

type svc struct {
}

func NewService() ISVC {
	return &svc{}
}

func (s *svc) List(namespace string) (*corev1.ServiceList, restErrors.IRestErr) {
	records := &corev1.ServiceList{}
	err := k8s.K8sClient.List(context.Background(), records, &client.ListOptions{Namespace: namespace}, &client.MatchingLabels{"app.kubernetes.io/managed-by": "kotal-operator"})
	if err != nil {
		go logger.Error(s.List, err)
		return nil, restErrors.NewInternalServerError(err.Error())
	}
	return records, nil
}

func (s *svc) Get(name string, namespace string) (*corev1.Service, restErrors.IRestErr) {
	record := &corev1.Service{}
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	err := k8s.K8sClient.Get(context.Background(), key, record)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, restErrors.NewNotFoundError(fmt.Sprintf("can't find service %s", name))
		}
		go logger.Error(s.Get, err)
		return nil, restErrors.NewInternalServerError(err.Error())
	}
	return record, nil
}
func (s *svc) Create(obj *corev1.Service) restErrors.IRestErr {
	err := k8s.K8sClient.Create(context.Background(), obj)
	if err != nil {
		if errors.IsConflict(err) {
			return restErrors.NewNotFoundError(fmt.Sprintf("service %s already exist", obj.Name))
		}
		go logger.Error(s.Get, err)
		return restErrors.NewInternalServerError(err.Error())
	}
	return nil
}
