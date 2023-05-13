package statefulset

import (
	"context"
	"github.com/kotalco/cloud-api/pkg/k8s"
	restError "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IStatefulSet interface {
	Count() (uint, *restError.RestErr)
	List(namespace string) (appsv1.StatefulSetList, *restError.RestErr)
}

type stateful struct {
}

func NewService() IStatefulSet {
	return &stateful{}
}

func (s *stateful) Count() (uint, *restError.RestErr) {
	list := &appsv1.StatefulSetList{}

	err := k8s.K8sClient.List(context.Background(), list, &client.MatchingLabels{"app.kubernetes.io/managed-by": "kotal-operator"})
	if err != nil {
		go logger.Error(s.Count, err)
		return 0, restError.NewInternalServerError("can't get stateful set count")
	}
	return uint(len(list.Items)), nil
}

func (s *stateful) List(namespace string) (stsList appsv1.StatefulSetList, restErr *restError.RestErr) {
	err := k8s.K8sClient.List(context.Background(), &stsList, &client.MatchingLabels{"app.kubernetes.io/managed-by": "kotal-operator"}, client.InNamespace(namespace))
	if err != nil {
		go logger.Error("DEPLOYMENT_COUNT", err)
		restErr = restError.NewInternalServerError(err.Error())
		return
	}
	return
}
