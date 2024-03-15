package deployment

import (
	"context"
	"github.com/kotalco/core-api/k8s"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
)

var k8sClient = k8s.NewClientService()

type IDeployment interface {
	Get(key types.NamespacedName) (*appsv1.Deployment, restErrors.IRestErr)
	Update(record *appsv1.Deployment) restErrors.IRestErr
}

type deployment struct {
}

func NewService() IDeployment {
	return &deployment{}
}

func (s *deployment) Get(key types.NamespacedName) (*appsv1.Deployment, restErrors.IRestErr) {
	record := &appsv1.Deployment{}
	err := k8sClient.Get(context.Background(), key, record)
	if err != nil {
		restErr := restErrors.NewNotFoundError(err.Error())
		return nil, restErr
	}
	return record, nil
}

func (s *deployment) Update(record *appsv1.Deployment) restErrors.IRestErr {
	err := k8sClient.Update(context.Background(), record)
	if err != nil {
		restErr := restErrors.NewInternalServerError(err.Error())
		return restErr
	}
	return nil
}
