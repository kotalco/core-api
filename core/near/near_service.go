package near

import (
	"context"
	"fmt"
	"github.com/kotalco/core-api/k8s"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/logger"
	nearv1alpha1 "github.com/kotalco/kotal/apis/near/v1alpha1"
	sharedAPIs "github.com/kotalco/kotal/apis/shared"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type nearService struct{}

type IService interface {
	Get(types.NamespacedName) (nearv1alpha1.Node, restErrors.IRestErr)
	Create(NearDto) (nearv1alpha1.Node, restErrors.IRestErr)
	Update(NearDto, *nearv1alpha1.Node) restErrors.IRestErr
	List(namespace string) (nearv1alpha1.NodeList, restErrors.IRestErr)
	Delete(*nearv1alpha1.Node) restErrors.IRestErr
	Count(namespace string) (int, restErrors.IRestErr)
}

var (
	k8sClient = k8s.NewClientService()
)

func NewNearService() IService {
	return nearService{}
}

// Get gets a single near node by name
func (service nearService) Get(namespacedName types.NamespacedName) (node nearv1alpha1.Node, restErr restErrors.IRestErr) {
	if err := k8sClient.Get(context.Background(), namespacedName, &node); err != nil {
		if apiErrors.IsNotFound(err) {
			restErr = restErrors.NewNotFoundError(fmt.Sprintf("node by name %s doesn't exit", namespacedName))
			return
		}
		go logger.Error(service.Get, err)
		restErr = restErrors.NewInternalServerError(fmt.Sprintf("can't get node by name %s", namespacedName))
		return
	}

	return
}

// Create creates near node from spec
func (service nearService) Create(dto NearDto) (node nearv1alpha1.Node, restErr restErrors.IRestErr) {
	node.ObjectMeta = dto.ObjectMetaFromMetadataDto()
	node.Spec = nearv1alpha1.NodeSpec{
		Network: dto.Network,
		Archive: dto.Archive,
		RPC:     true,
		Image:   dto.Image,
		Resources: sharedAPIs.Resources{
			StorageClass: dto.StorageClass,
		},
	}

	k8s.DefaultResources(&node.Spec.Resources)

	if os.Getenv("MOCK") == "true" {
		node.Default()
	}

	if err := k8sClient.Create(context.Background(), &node); err != nil {
		if apiErrors.IsAlreadyExists(err) {
			restErr = restErrors.NewNotFoundError(fmt.Sprintf("node by name %s already exits", node.Name))
			return
		}
		go logger.Error(service.Create, err)
		restErr = restErrors.NewInternalServerError("failed to create node")
		return
	}

	return
}

// Update updates near node by name from spec
func (service nearService) Update(dto NearDto, node *nearv1alpha1.Node) (restErr restErrors.IRestErr) {

	if dto.NodePrivateKeySecretName != nil {
		node.Spec.NodePrivateKeySecretName = *dto.NodePrivateKeySecretName
	}

	if dto.ValidatorSecretName != nil {
		node.Spec.ValidatorSecretName = *dto.ValidatorSecretName
	}

	if dto.MinPeers != 0 {
		node.Spec.MinPeers = dto.MinPeers
	}

	if dto.P2PPort != 0 {
		node.Spec.P2PPort = dto.P2PPort
	}

	if dto.RPC != nil {
		node.Spec.RPC = *dto.RPC
	}
	if node.Spec.RPC {
		if dto.RPCPort != 0 {
			node.Spec.RPCPort = dto.RPCPort
		}
	}

	if dto.PrometheusPort != 0 {
		node.Spec.PrometheusPort = dto.PrometheusPort
	}

	if dto.TelemetryURL != nil {
		node.Spec.TelemetryURL = *dto.TelemetryURL
	}

	if bootnodes := dto.Bootnodes; bootnodes != nil {
		node.Spec.Bootnodes = *bootnodes
	}

	if dto.CPU != "" {
		node.Spec.CPU = dto.CPU
	}
	if dto.CPULimit != "" {
		node.Spec.CPULimit = dto.CPULimit
	}
	if dto.Memory != "" {
		node.Spec.Memory = dto.Memory
	}
	if dto.MemoryLimit != "" {
		node.Spec.MemoryLimit = dto.MemoryLimit
	}
	if dto.Storage != "" {
		node.Spec.Storage = dto.Storage
	}
	if dto.Image != "" {
		node.Spec.Image = dto.Image
	}

	if os.Getenv("MOCK") == "true" {
		node.Default()
	}

	pod := &corev1.Pod{}
	podIsPending := false
	if dto.CPU != "" || dto.Memory != "" {
		key := types.NamespacedName{
			Namespace: node.Namespace,
			Name:      fmt.Sprintf("%s-0", node.Name),
		}
		err := k8sClient.Get(context.Background(), key, pod)
		if apiErrors.IsNotFound(err) {
			go logger.Error(service.Update, err)
			restErr = restErrors.NewBadRequestError(fmt.Sprintf("pod by name %s doesn't exit", key.Name))
			return
		}
		podIsPending = pod.Status.Phase == corev1.PodPending
	}

	if err := k8sClient.Update(context.Background(), node); err != nil {
		go logger.Error(service.Update, err)
		restErr = restErrors.NewInternalServerError(fmt.Sprintf("can't update node by name %s", node.Name))
		return
	}

	if podIsPending {
		err := k8sClient.Delete(context.Background(), pod)
		if err != nil {
			go logger.Error(service.Update, err)
			restErr = restErrors.NewInternalServerError(fmt.Sprintf("can't update node by name %s", node.Name))
			return
		}
	}

	return
}

// List returns all near nodes
func (service nearService) List(namespace string) (list nearv1alpha1.NodeList, restErr restErrors.IRestErr) {
	if err := k8sClient.List(context.Background(), &list, client.InNamespace(namespace)); err != nil {
		go logger.Error(service.List, err)
		restErr = restErrors.NewInternalServerError("failed to get all nodes")
		return
	}

	return
}

// Count returns total number of near nodes
func (service nearService) Count(namespace string) (count int, restErr restErrors.IRestErr) {
	nodes := &nearv1alpha1.NodeList{}
	if err := k8sClient.List(context.Background(), nodes, client.InNamespace(namespace)); err != nil {
		go logger.Error(service.Count, err)
		restErr = restErrors.NewInternalServerError("failed to count all nodes")
		return
	}

	return len(nodes.Items), nil
}

// Delete deletes ethereum 2.0 near node by name
func (service nearService) Delete(node *nearv1alpha1.Node) (restErr restErrors.IRestErr) {
	if err := k8sClient.Delete(context.Background(), node); err != nil {
		go logger.Error(service.Delete, err)
		restErr = restErrors.NewInternalServerError(fmt.Sprintf("can't delete node by name %s", node.Name))
		return
	}

	return
}
