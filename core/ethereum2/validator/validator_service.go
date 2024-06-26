package validator

import (
	"context"
	"fmt"
	"github.com/kotalco/core-api/k8s"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/logger"
	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	sharedAPIs "github.com/kotalco/kotal/apis/shared"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type validatorService struct{}

type IService interface {
	Get(types.NamespacedName) (ethereum2v1alpha1.Validator, restErrors.IRestErr)
	Create(dto ValidatorDto) (ethereum2v1alpha1.Validator, restErrors.IRestErr)
	Update(ValidatorDto, *ethereum2v1alpha1.Validator) restErrors.IRestErr
	List(namespace string) (ethereum2v1alpha1.ValidatorList, restErrors.IRestErr)
	Delete(*ethereum2v1alpha1.Validator) restErrors.IRestErr
	Count(namespace string) (int, restErrors.IRestErr)
}

var (
	k8sClient = k8s.NewClientService()
)

func NewValidatorService() IService {
	return validatorService{}
}

// Get gets a single ethereum 2.0 beacon node by name
func (service validatorService) Get(namespacedName types.NamespacedName) (validator ethereum2v1alpha1.Validator, restErr restErrors.IRestErr) {
	if err := k8sClient.Get(context.Background(), namespacedName, &validator); err != nil {
		if apiErrors.IsNotFound(err) {
			restErr = restErrors.NewNotFoundError(fmt.Sprintf("validator by name %s doesn't exit", namespacedName.Name))
			return
		}
		go logger.Error(service.Get, err)
		restErr = restErrors.NewInternalServerError(fmt.Sprintf("can't get a validator by name %s", namespacedName.Name))
		return
	}

	return
}

// Create creates ethereum 2.0 beacon node from spec
func (service validatorService) Create(dto ValidatorDto) (validator ethereum2v1alpha1.Validator, restErr restErrors.IRestErr) {
	validator.ObjectMeta = dto.ObjectMetaFromMetadataDto()
	validator.Spec = ethereum2v1alpha1.ValidatorSpec{
		Network:   dto.Network,
		Client:    ethereum2v1alpha1.Ethereum2Client(dto.Client),
		Keystores: dto.Keystores,
		Image:     dto.Image,
		Resources: sharedAPIs.Resources{
			StorageClass: dto.StorageClass,
		},
	}

	k8s.DefaultResources(&validator.Spec.Resources)

	if dto.Client == string(ethereum2v1alpha1.PrysmClient) && dto.WalletPasswordSecretName != "" {
		validator.Spec.WalletPasswordSecret = dto.WalletPasswordSecretName
	}

	if len(dto.BeaconEndpoints) != 0 {
		validator.Spec.BeaconEndpoints = dto.BeaconEndpoints
	} else {
		validator.Spec.BeaconEndpoints = []string{}
	}

	if os.Getenv("MOCK") == "true" {
		validator.Default()
	}

	if err := k8sClient.Create(context.Background(), &validator); err != nil {
		if apiErrors.IsAlreadyExists(err) {
			restErr = restErrors.NewNotFoundError(fmt.Sprintf("validator by name %s already exits", validator.Name))
			return
		}
		go logger.Error(service.Create, err)
		restErr = restErrors.NewInternalServerError("failed to create validator")
		return
	}

	return
}

// Update updates ethereum 2.0 beacon node by name from spec
func (service validatorService) Update(dto ValidatorDto, validator *ethereum2v1alpha1.Validator) (restErr restErrors.IRestErr) {
	if dto.WalletPasswordSecretName != "" {
		validator.Spec.WalletPasswordSecret = dto.WalletPasswordSecretName
	}

	if len(dto.Keystores) != 0 {
		keystores := []ethereum2v1alpha1.Keystore{}
		for _, keystore := range dto.Keystores {
			keystores = append(keystores, ethereum2v1alpha1.Keystore{
				SecretName: keystore.SecretName,
			})
		}
		validator.Spec.Keystores = keystores
	}

	if dto.Graffiti != "" {
		validator.Spec.Graffiti = dto.Graffiti
	}

	if len(dto.BeaconEndpoints) != 0 {
		validator.Spec.BeaconEndpoints = dto.BeaconEndpoints
	}

	if dto.CPU != "" {
		validator.Spec.CPU = dto.CPU
	}
	if dto.CPULimit != "" {
		validator.Spec.CPULimit = dto.CPULimit
	}
	if dto.Memory != "" {
		validator.Spec.Memory = dto.Memory
	}
	if dto.MemoryLimit != "" {
		validator.Spec.MemoryLimit = dto.MemoryLimit
	}
	if dto.Storage != "" {
		validator.Spec.Storage = dto.Storage
	}
	if dto.Image != "" {
		validator.Spec.Image = dto.Image
	}

	if os.Getenv("MOCK") == "true" {
		validator.Default()
	}

	pod := &corev1.Pod{}
	podIsPending := false
	if dto.CPU != "" || dto.Memory != "" {
		key := types.NamespacedName{
			Namespace: validator.Namespace,
			Name:      fmt.Sprintf("%s-0", validator.Name),
		}
		err := k8sClient.Get(context.Background(), key, pod)
		if apiErrors.IsNotFound(err) {
			go logger.Error(service.Update, err)
			restErr = restErrors.NewBadRequestError(fmt.Sprintf("pod by name %s doesn't exit", key.Name))
			return
		}
		podIsPending = pod.Status.Phase == corev1.PodPending
	}

	if err := k8sClient.Update(context.Background(), validator); err != nil {
		go logger.Error(service.Update, err)
		restErr = restErrors.NewInternalServerError(fmt.Sprintf("can't update node by name %s", validator.Name))
		return
	}

	if podIsPending {
		err := k8sClient.Delete(context.Background(), pod)
		if err != nil {
			go logger.Error(service.Update, err)
			restErr = restErrors.NewInternalServerError(fmt.Sprintf("can't update node by name %s", validator.Name))
			return
		}
	}
	return
}

// List returns all ethereum 2.0 beacon nodes
func (service validatorService) List(namespace string) (list ethereum2v1alpha1.ValidatorList, restErr restErrors.IRestErr) {
	if err := k8sClient.List(context.Background(), &list, client.InNamespace(namespace)); err != nil {
		go logger.Error(service.List, err)
		restErr = restErrors.NewInternalServerError("failed to get all validators")
		return
	}

	return
}

// Count returns total number of beacon nodes
func (service validatorService) Count(namespace string) (count int, restErr restErrors.IRestErr) {
	validators := &ethereum2v1alpha1.ValidatorList{}

	if err := k8sClient.List(context.Background(), validators, client.InNamespace(namespace)); err != nil {
		go logger.Error(service.Count, err)
		restErr = restErrors.NewInternalServerError("error counting validators")
		return
	}

	return len(validators.Items), nil
}

// Delete deletes ethereum 2.0 beacon node by name
func (service validatorService) Delete(validator *ethereum2v1alpha1.Validator) (restErr restErrors.IRestErr) {
	if err := k8sClient.Delete(context.Background(), validator); err != nil {
		go logger.Error(service.Delete, err)
		restErr = restErrors.NewBadRequestError(fmt.Sprintf("can't delete validator by name %s", validator.Name))
		return
	}

	return
}
