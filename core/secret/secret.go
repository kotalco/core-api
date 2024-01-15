package secret

import (
	"github.com/kotalco/core-api/k8s"
	"github.com/kotalco/core-api/pkg/time"
	corev1 "k8s.io/api/core/v1"
)

type SecretDto struct {
	time.Time
	k8s.MetaDataDto
	Type string            `json:"type"`
	Data map[string]string `json:"data,omitempty"`
}

type SecretsDto []SecretDto

func (dto SecretDto) FromCoreSecret(s corev1.Secret) SecretDto {
	dto.Name = s.Name
	dto.Time = time.Time{CreatedAt: s.CreationTimestamp.UTC().Format(time.JavascriptISOString)}
	dto.Type = s.Labels["kotal.io/key-type"]

	return dto
}

func (secret SecretsDto) FromCoreSecret(secrets []corev1.Secret) SecretsDto {
	result := make(SecretsDto, len(secrets))
	for index, value := range secrets {
		result[index] = SecretDto{}.FromCoreSecret(value)
	}
	return result
}
