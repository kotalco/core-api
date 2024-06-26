package validator

import (
	"github.com/kotalco/core-api/k8s"
	"github.com/kotalco/core-api/pkg/time"
	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	sharedAPI "github.com/kotalco/kotal/apis/shared"
)

type ValidatorDto struct {
	time.Time
	k8s.MetaDataDto
	Network                  string                       `json:"network"`
	Client                   string                       `json:"client"`
	Graffiti                 string                       `json:"graffiti"`
	BeaconEndpoints          []string                     `json:"beaconEndpoints"`
	WalletPasswordSecretName string                       `json:"walletPasswordSecretName"`
	Keystores                []ethereum2v1alpha1.Keystore `json:"keystores"`
	Image                    string                       `json:"image"`
	sharedAPI.Resources
}

type ValidatorListDto []ValidatorDto

func (dto ValidatorDto) FromEthereum2Validator(validator ethereum2v1alpha1.Validator) ValidatorDto {
	dto.Name = validator.Name
	dto.Time = time.Time{CreatedAt: validator.CreationTimestamp.UTC().Format(time.JavascriptISOString)}
	dto.Network = validator.Spec.Network
	dto.Client = string(validator.Spec.Client)
	dto.Graffiti = validator.Spec.Graffiti
	dto.BeaconEndpoints = validator.Spec.BeaconEndpoints
	dto.Keystores = validator.Spec.Keystores
	dto.CPU = validator.Spec.CPU
	dto.CPULimit = validator.Spec.CPULimit
	dto.Memory = validator.Spec.Memory
	dto.MemoryLimit = validator.Spec.MemoryLimit
	dto.Storage = validator.Spec.Storage
	dto.StorageClass = validator.Spec.StorageClass
	dto.WalletPasswordSecretName = validator.Spec.WalletPasswordSecret
	dto.Image = validator.Spec.Image

	return dto
}

func (validatorListDto ValidatorListDto) FromEthereum2Validator(validators []ethereum2v1alpha1.Validator) ValidatorListDto {
	result := make(ValidatorListDto, len(validators))
	for index, v := range validators {
		result[index] = ValidatorDto{}.FromEthereum2Validator(v)
	}
	return result
}
