// Package secret internal is the domain layer for creating secrets
// uses the k8 client to CRUD the node
package secret

import (
	"context"
	"fmt"
	"github.com/kotalco/core-api/k8s"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type secretService struct{}

type IService interface {
	Get(name types.NamespacedName) (corev1.Secret, restErrors.IRestErr)
	Create(SecretDto) (corev1.Secret, restErrors.IRestErr)
	List(namespace string) (corev1.SecretList, restErrors.IRestErr)
	Delete(*corev1.Secret) restErrors.IRestErr
	Count(namespace string) (int, restErrors.IRestErr)
}

var (
	k8sClient = k8s.NewClientService()
)

func NewSecretService() IService {
	return secretService{}
}

// Get returns a single secret  by name
func (service secretService) Get(namespacedName types.NamespacedName) (secret corev1.Secret, restErr restErrors.IRestErr) {
	if err := k8sClient.Get(context.Background(), namespacedName, &secret); err != nil {
		if apiErrors.IsNotFound(err) {
			restErr = restErrors.NewNotFoundError(fmt.Sprintf("secret by name %s doesn't exist", namespacedName.Name))
			return
		}
		go logger.Error(service.Get, err)
		restErr = restErrors.NewInternalServerError(fmt.Sprintf("can't get secret by name %s", namespacedName.Name))
		return
	}
	return
}

// Create creates a secret from the given spec
func (service secretService) Create(dto SecretDto) (secret corev1.Secret, restErr restErrors.IRestErr) {

	t := true
	secret.ObjectMeta = metav1.ObjectMeta{
		Name:      dto.Name,
		Namespace: dto.Namespace,
		Labels: map[string]string{
			"kotal.io/key-type":            dto.Type,
			"app.kubernetes.io/created-by": "kotal-api",
		},
	}
	secret.StringData = dto.Data
	secret.Immutable = &t

	if err := k8sClient.Create(context.Background(), &secret); err != nil {
		if apiErrors.IsAlreadyExists(err) {
			restErr = restErrors.NewBadRequestError(fmt.Sprintf("secret by name %s already exist", dto.Name))
			return
		}
		go logger.Error(service.Create, err)
		restErr = restErrors.NewInternalServerError("error creating secret")
		return
	}
	return
}

// List returns all secrets
func (service secretService) List(namespace string) (list corev1.SecretList, restErr restErrors.IRestErr) {
	if err := k8sClient.List(context.Background(), &list, client.InNamespace(namespace), client.HasLabels{"app.kubernetes.io/created-by"}); err != nil {
		go logger.Error(service.List, err)
		restErr = restErrors.NewInternalServerError("failed to get all secrets")
		return
	}

	return
}

// Delete a single secret node by name
func (service secretService) Delete(secret *corev1.Secret) (restErr restErrors.IRestErr) {
	if err := k8sClient.Delete(context.Background(), secret); err != nil {
		go logger.Error(service.Delete, err)
		restErr = restErrors.NewInternalServerError(fmt.Sprintf("can't delete secret by name %s", secret.Name))
		return
	}
	return
}

// Count counts secrets
func (service secretService) Count(namespace string) (count int, restErr restErrors.IRestErr) {
	secrets := &corev1.SecretList{}
	if err := k8sClient.List(context.Background(), secrets, client.InNamespace(namespace), client.HasLabels{"kotal.io/key-type"}); err != nil {
		go logger.Error(service.Count, err)
		restErr = restErrors.NewInternalServerError("failed to get all secrets")
		return
	}
	return len(secrets.Items), nil
}
