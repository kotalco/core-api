package svc

import (
	"context"
	"github.com/kotalco/cloud-api/pkg/k8s"
	"github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ISVC interface {
	List(namespace string) (*corev1.ServiceList, *errors.RestErr)
	Get(name string, namespace string) (*corev1.Service, *errors.RestErr)
}

type svc struct {
}

func NewService() ISVC {
	return &svc{}
}

func (s *svc) List(namespace string) (*corev1.ServiceList, *errors.RestErr) {
	records := &corev1.ServiceList{}
	err := k8s.K8sClient.List(context.Background(), records, &client.ListOptions{Namespace: namespace}, &client.MatchingLabels{"app.kubernetes.io/managed-by": "kotal"})
	if err != nil {
		go logger.Error(s.List, err)
		return nil, errors.NewInternalServerError(err.Error())
	}
	return records, nil
}

func (s *svc) Get(name string, namespace string) (*corev1.Service, *errors.RestErr) {
	record := &corev1.Service{}
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	err := k8s.K8sClient.Get(context.Background(), key, record)
	if err != nil {
		go logger.Error(s.Get, err)
		return nil, errors.NewInternalServerError(err.Error())
	}
	return record, nil
}