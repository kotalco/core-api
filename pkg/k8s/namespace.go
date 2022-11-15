package k8s

import (
	"context"
	"fmt"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// namespace service communicates with k8s from community-api to create namespaces for different nodes
// in combination with workspace-module they form (workspace exposed to the user which is the namespace behind the scenes)
type namespace struct{}

type INamespace interface {
	Create(name string) *restErrors.RestErr
	Get(name string) (*corev1.Namespace, *restErrors.RestErr)
	Delete(name string) *restErrors.RestErr
}

// NewNamespaceService returns new instance of the namespace service
func NewNamespaceService() INamespace {
	newNamespace := &namespace{}
	return newNamespace
}

// Create creates new namespace from a given name using the clientSet from community-api microservice
func (service *namespace) Create(name string) *restErrors.RestErr {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{"app.kubernetes.io/created-by": "kotal"},
		},
	}
	err := K8sClient.Create(context.Background(), ns)
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

// Get returns a namespace if exits
func (service *namespace) Get(name string) (*corev1.Namespace, *restErrors.RestErr) {
	ns := corev1.Namespace{}

	key := types.NamespacedName{}
	key.Namespace = name
	key.Name = name

	err := K8sClient.Get(context.Background(), key, &ns)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, restErrors.NewNotFoundError(fmt.Sprintf("can't find namespace %s", name))
		}
		logger.Error(service.Get, err)
		return nil, restErrors.NewInternalServerError("something went wrong")
	}
	return &ns, nil
}

// Delete deletes a namespace if exits
func (service *namespace) Delete(name string) *restErrors.RestErr {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	err := K8sClient.Delete(context.Background(), ns)
	if err != nil {
		if errors.IsNotFound(err) {
			return restErrors.NewNotFoundError(fmt.Sprintf("namespace %s does't exit", name))
		}
		logger.Error(service.Delete, err)
		return restErrors.NewInternalServerError("something went wrong!")
	}
	return nil
}
