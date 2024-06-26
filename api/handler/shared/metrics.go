package shared

import (
	"context"
	"fmt"
	"github.com/kotalco/core-api/pkg/responder"
	"time"

	restError "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gofiber/websocket/v2"
	"github.com/kotalco/core-api/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

var metricsClientset = k8s.MetricsClientset()

type metricsResponseDto struct {
	Cpu    int64 `json:"cpu"`
	Memory int64 `json:"memory"`
}

// Metrics returns a websocket that emits cpu and memory usage
func Metrics(c *websocket.Conn) {
	defer c.Close()

	name := c.Params("name")
	ns := c.Locals("namespace").(string)
	pod := &corev1.Pod{}

	key := types.NamespacedName{
		Namespace: ns,
		Name:      fmt.Sprintf("%s-0", name),
	}

	sts := &appsv1.StatefulSet{}
	stsKey := types.NamespacedName{
		Namespace: ns,
		Name:      name,
	}

podCheck:
	if err := k8sClient.Get(context.Background(), key, pod); err != nil {
		go logger.Info("METRICS_POD_NOTFOUND", err.Error())
		// is the pod error due to sts has been deleted ?
		stsErr := k8sClient.Get(context.Background(), stsKey, sts)
		if apierrors.IsNotFound(stsErr) {
			go logger.Info("METRICS_STS_NOTFOUND", stsErr.Error())
			c.WriteJSON(responder.NewResponse(restError.NewNotFoundError(stsErr.Error())))
			return
		}
		time.Sleep(3 * time.Second)
		goto podCheck
	}

	opts := metav1.GetOptions{}
	podMetrics := metricsClientset.MetricsV1beta1().PodMetricses(key.Namespace)

	for {
		response := new(metricsResponseDto)

		metrics, err := podMetrics.Get(context.Background(), key.Name, opts)
		if err != nil {
			go logger.Info("METRICS_API_ERR", err.Error())
			time.Sleep(3 * time.Second)
			goto podCheck
		}

		response.Cpu = metrics.Containers[0].Usage.Cpu().ScaledValue(resource.Milli)
		response.Memory = metrics.Containers[0].Usage.Memory().ScaledValue(resource.Mega)

		if err := c.WriteJSON(response); err != nil {
			return
		}

		time.Sleep(time.Second)
	}
}
