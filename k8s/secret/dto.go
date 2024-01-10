package secret

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CreateSecretDto struct {
	metav1.ObjectMeta
	Type       v1.SecretType
	StringData map[string]string
	OwnersRef  []metav1.OwnerReference
}
