package kotal_traefik

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
	SetLetsEncryptStaticConfiguration(resolverName string, acmeEmail string) restErrors.IRestErr
	DeleteLetsEncryptStaticConfiguration() restErrors.IRestErr
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

func (s *traefik) SetLetsEncryptStaticConfiguration(resolverNme string, acmeEmail string) restErrors.IRestErr {
	deploy, restErr := s.Get()
	if restErr != nil {
		return restErr
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
		go logger.Warn(s.SetLetsEncryptStaticConfiguration, err)
		return restErrors.NewInternalServerError(err.Error())
	}
	return nil
}

func (s *traefik) DeleteLetsEncryptStaticConfiguration() restErrors.IRestErr {
	deploy, restErr := s.Get()
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
		go logger.Warn(s.SetLetsEncryptStaticConfiguration, err)
		return restErrors.NewInternalServerError(err.Error())
	}
	return nil
}
