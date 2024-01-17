package statefulset

import (
	"context"
	"github.com/kotalco/core-api/k8s"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var k8sClient = k8s.NewClientService()

type IStatefulSet interface {
	Count() (uint, restErrors.IRestErr)
	List(namespace string) (appsv1.StatefulSetList, restErrors.IRestErr)
}

type stateful struct {
}

func NewService() IStatefulSet {
	return &stateful{}
}

func (s *stateful) Count() (uint, restErrors.IRestErr) {
	list := &appsv1.StatefulSetList{}

	err := k8sClient.List(context.Background(), list, &client.MatchingLabels{"app.kubernetes.io/managed-by": "kotal-operator"})
	if err != nil {
		go logger.Error(s.Count, err)
		return 0, restErrors.NewInternalServerError("can't get stateful set count")
	}
	return uint(len(list.Items)), nil
}

func (s *stateful) List(namespace string) (stsList appsv1.StatefulSetList, restErr restErrors.IRestErr) {
	err := k8sClient.List(context.Background(), &stsList, &client.MatchingLabels{"app.kubernetes.io/managed-by": "kotal-operator"}, client.InNamespace(namespace))
	if err != nil {
		go logger.Error("DEPLOYMENT_COUNT", err)
		restErr = restErrors.NewInternalServerError(err.Error())
		return
	}
	return
}
