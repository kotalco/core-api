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
}

type stateful struct {
}

func NewService() IStatefulSet {
	return &stateful{}
}

func (s *stateful) Count() (uint, *restError.RestErr) {
	list := &appsv1.StatefulSetList{}

	err := k8s.K8sClient.List(context.Background(), list, &client.MatchingLabels{"app.kubernetes.io/managed-by": "kotal"})
	if err != nil {
		go logger.Error(s.Count, err)
		return 0, restError.NewInternalServerError("can't get stateful set count")
	}
	return uint(len(list.Items)), nil
}
