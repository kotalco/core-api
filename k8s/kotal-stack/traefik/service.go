package traefik

import (
	"context"
	"fmt"
	"github.com/kotalco/core-api/config"
	"github.com/kotalco/core-api/k8s"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"strings"
)

var k8sClient = k8s.NewClientService()

type ITraefik interface {
	Get() (*appsv1.Deployment, restErrors.IRestErr)
	SetLetsEncryptStaticConfiguration(deploy *appsv1.Deployment, acmeEmail string) restErrors.IRestErr
	DeleteLetsEncryptStaticConfiguration(deploy *appsv1.Deployment) restErrors.IRestErr
}

type traefik struct {
}

func NewService() ITraefik {

	return &traefik{}
}

func (s *traefik) Get() (*appsv1.Deployment, restErrors.IRestErr) {
	key := types.NamespacedName{Name: config.Environment.TraefikDeploymentName, Namespace: config.Environment.TraefikNamespace}
	record := &appsv1.Deployment{}
	err := k8sClient.Get(context.Background(), key, record)
	if err != nil {
		restErr := restErrors.NewNotFoundError(err.Error())
		return nil, restErr
	}
	return record, nil
}

func (s *traefik) SetLetsEncryptStaticConfiguration(deploy *appsv1.Deployment, acmeEmail string) restErrors.IRestErr {
	for i, container := range deploy.Spec.Template.Spec.Containers {
		if container.Name == config.Environment.TraefikDeploymentName {
			newArgs := container.Args
			newArgs = append(container.Args, "--certificatesresolvers.kotalletsresolver.acme.tlschallenge")
			newArgs = append(container.Args, fmt.Sprintf("--certificatesresolvers.kotalletsresolver.acme.email=%s", acmeEmail))
			newArgs = append(container.Args, "--certificatesresolvers.kotalletsresolver.acme.storage=/data/acme.json")
			deploy.Spec.Template.Spec.Containers[i].Args = newArgs
			break
		}
	}

	err := k8sClient.Update(context.Background(), deploy)
	if err != nil {
		go logger.Warn(s.SetLetsEncryptStaticConfiguration, err)
		return restErrors.NewInternalServerError(err.Error())
	}
	return nil
}

func (s *traefik) DeleteLetsEncryptStaticConfiguration(deploy *appsv1.Deployment) restErrors.IRestErr {
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
		go logger.Warn(s.SetLetsEncryptStaticConfiguration, err)
		return restErrors.NewInternalServerError(err.Error())
	}
	return nil
}
