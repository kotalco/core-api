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
)

var k8sClient = k8s.NewClientService()

type ISecret interface {
	Create(dto *CreateSecretDto) restErrors.IRestErr
	Get(name string, namespace string) (*corev1.Secret, restErrors.IRestErr)
}

type secret struct {
}

func NewService() ISecret {
	return &secret{}
}

func (s *secret) Create(dto *CreateSecretDto) restErrors.IRestErr {
	t := true
	secret := &corev1.Secret{
		Type: dto.Type,
		ObjectMeta: metav1.ObjectMeta{
			Name:            dto.Name,
			Namespace:       dto.Namespace,
			OwnerReferences: dto.OwnerReferences,
		},
		StringData: dto.StringData,
		Immutable:  &t,
	}

	if err := k8sClient.Create(context.Background(), secret); err != nil {
		if apiErrors.IsAlreadyExists(err) {
			return restErrors.NewBadRequestError(fmt.Sprintf("secret by name %s already exist", dto.Name))
		}
		go logger.Error(s.Create, err)
		return restErrors.NewInternalServerError("error creating secret")
	}
	return nil
}

func (s *secret) Get(name string, namespace string) (*corev1.Secret, restErrors.IRestErr) {
	record := &corev1.Secret{}
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	err := k8sClient.Get(context.Background(), key, record)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return nil, restErrors.NewNotFoundError(fmt.Sprintf("can't find secret %s", name))
		}
		go logger.Error(s.Get, err)
		return nil, restErrors.NewInternalServerError(err.Error())
	}
	return record, nil
}
