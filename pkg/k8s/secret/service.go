package secret

import (
	"context"
	"fmt"
	"github.com/kotalco/cloud-api/pkg/k8s"
	restError "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type ISecret interface {
	Create(dto *CreateSecretDto) *restError.RestErr
	Get(name string, namespace string) (*corev1.Secret, *restError.RestErr)
}

type secret struct {
}

func NewService() ISecret {
	return &secret{}
}

func (s *secret) Create(dto *CreateSecretDto) *restError.RestErr {
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

	if err := k8s.K8sClient.Create(context.Background(), secret); err != nil {
		if apiErrors.IsAlreadyExists(err) {
			return restError.NewBadRequestError(fmt.Sprintf("secret by name %s already exist", dto.Name))
		}
		go logger.Error(s.Create, err)
		return restError.NewInternalServerError("error creating secret")
	}
	return nil
}

func (s *secret) Get(name string, namespace string) (*corev1.Secret, *restError.RestErr) {
	record := &corev1.Secret{}
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	err := k8s.K8sClient.Get(context.Background(), key, record)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return nil, restError.NewNotFoundError(fmt.Sprintf("can't find secret %s", name))
		}
		go logger.Error(s.Get, err)
		return nil, restError.NewInternalServerError(err.Error())
	}
	return record, nil
}
