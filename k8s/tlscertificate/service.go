package tlscertificate

import (
	"context"
	"fmt"
	"github.com/kotalco/core-api/config"
	"github.com/kotalco/core-api/core/setting"
	"github.com/kotalco/core-api/k8s"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/logger"
	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	"github.com/traefik/traefik/v2/pkg/tls"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"strings"
)

var k8sClient = k8s.NewClientService()

type TLSCertificate interface {
	GetTraefikDeployment() (*appsv1.Deployment, restErrors.IRestErr)
	ConfigureLetsEncrypt(resolverNme string, acmeEmail string) restErrors.IRestErr
	ConfigureCustomCertificate(secretName string) restErrors.IRestErr
}

type tlsCertificate struct{}

func NewService() TLSCertificate { return &tlsCertificate{} }

func (t *tlsCertificate) GetTraefikDeployment() (*appsv1.Deployment, restErrors.IRestErr) {
	key := types.NamespacedName{Name: config.Environment.TraefikDeploymentName, Namespace: config.Environment.TraefikNamespace}
	record := &appsv1.Deployment{}
	err := k8sClient.Get(context.Background(), key, record)
	if err != nil {
		restErr := restErrors.NewNotFoundError(err.Error())
		return nil, restErr
	}
	return record, nil
}

func (t *tlsCertificate) ConfigureLetsEncrypt(resolverNme string, acmeEmail string) restErrors.IRestErr {
	//delete default tls-store if exists
	tlsStore := &traefikv1alpha1.TLSStore{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: config.Environment.TraefikNamespace,
		},
	}
	_ = k8sClient.Delete(context.Background(), tlsStore)

	//create tls store
	tlsStore = &traefikv1alpha1.TLSStore{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: config.Environment.TraefikNamespace,
		},
		Spec: traefikv1alpha1.TLSStoreSpec{
			DefaultGeneratedCert: &tls.GeneratedCert{Resolver: setting.KotalLetsEncryptResolverName},
		},
	}
	_ = k8sClient.Create(context.Background(), tlsStore)

	deploy, restErr := t.GetTraefikDeployment()
	if restErr != nil {
		return restErr
	}

	for i, container := range deploy.Spec.Template.Spec.Containers {
		if container.Name == config.Environment.TraefikDeploymentName {
			var newArgs []string
			for _, arg := range container.Args {
				if !strings.Contains(arg, "certificatesresolvers") {
					newArgs = append(newArgs, arg)
				}
			}
			deploy.Spec.Template.Spec.Containers[i].Args = newArgs
			break
		}
	}

	for i, container := range deploy.Spec.Template.Spec.Containers {
		if container.Name == config.Environment.TraefikDeploymentName {
			container.Args = append(container.Args, fmt.Sprintf("--certificatesresolvers.%s.acme.tlschallenge", resolverNme))
			container.Args = append(container.Args, fmt.Sprintf("--certificatesresolvers.%s.acme.email=%s", resolverNme, acmeEmail))
			container.Args = append(container.Args, fmt.Sprintf("--certificatesresolvers.%s.acme.storage=/data/acme.json", resolverNme))
			deploy.Spec.Template.Spec.Containers[i].Args = container.Args
			break
		}
	}

	err := k8sClient.Update(context.Background(), deploy)
	if err != nil {
		go logger.Warn(t.ConfigureLetsEncrypt, err)
		return restErrors.NewInternalServerError(err.Error())
	}
	return nil
}

func (t *tlsCertificate) ConfigureCustomCertificate(secretName string) restErrors.IRestErr {
	//delete default tls-store if exists
	tlsStore := &traefikv1alpha1.TLSStore{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: config.Environment.TraefikNamespace,
		},
	}
	_ = k8sClient.Delete(context.Background(), tlsStore)

	//create tls store
	tlsStore = &traefikv1alpha1.TLSStore{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: config.Environment.TraefikNamespace,
		},
		Spec: traefikv1alpha1.TLSStoreSpec{
			DefaultCertificate: &traefikv1alpha1.Certificate{SecretName: secretName},
		},
	}
	_ = k8sClient.Create(context.Background(), tlsStore)

	//remove letsEncrypt static configuration
	deploy, restErr := t.GetTraefikDeployment()
	if restErr != nil {
		return restErr
	}

	for i, container := range deploy.Spec.Template.Spec.Containers {
		if container.Name == config.Environment.TraefikDeploymentName {
			var newArgs []string
			for _, arg := range container.Args {
				if !strings.Contains(arg, "certificatesresolvers") {
					newArgs = append(newArgs, arg)
				}
			}
			deploy.Spec.Template.Spec.Containers[i].Args = newArgs
			break
		}
	}

	err := k8sClient.Update(context.Background(), deploy)
	if err != nil {
		go logger.Warn(t.ConfigureCustomCertificate, err)
		return restErrors.NewInternalServerError(err.Error())
	}

	return nil
}
