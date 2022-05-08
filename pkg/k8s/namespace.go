package k8s

import (
	"context"
	"fmt"
	restErrors "github.com/kotalco/api/pkg/errors"
	communityK8s "github.com/kotalco/api/pkg/k8s"
	"github.com/kotalco/api/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type namespace struct{}

type INamespace interface {
	Create(name string) *restErrors.RestErr
	Get(name string) (*corev1.Namespace, *restErrors.RestErr)
	Delete(name string) *restErrors.RestErr
}

func NewNamespace() INamespace {
	newNamespace := &namespace{}
	return newNamespace
}

var clientSet = communityK8s.Clientset()

func (service *namespace) Create(name string) *restErrors.RestErr {
	nsName := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	_, err := clientSet.CoreV1().Namespaces().Create(context.Background(), nsName, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return restErrors.NewConflictError("namespace already exits")
		} else if errors.IsInvalid(err) {
			return restErrors.NewBadRequestError("invalid namespace value")
		}
		logger.Error(service.Create, err)
		return restErrors.NewInternalServerError("can't create namespace")
	}
	return nil
}

func (service *namespace) Get(name string) (*corev1.Namespace, *restErrors.RestErr) {
	workspace, err := clientSet.CoreV1().Namespaces().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, restErrors.NewNotFoundError(fmt.Sprintf("can't find namespace %s", name))
		}
		logger.Error(service.Get, err)
		return nil, restErrors.NewInternalServerError("something went wrong")
	}
	return workspace, nil
}

func (service *namespace) Delete(name string) *restErrors.RestErr {
	err := clientSet.CoreV1().Namespaces().Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return restErrors.NewNotFoundError(fmt.Sprintf("namespace %s  does't exit", name))
		}
		logger.Error(service.Delete, err)
		return restErrors.NewInternalServerError("something went wrong!")
	}
	return nil
}
