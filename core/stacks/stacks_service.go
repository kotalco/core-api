package stacks

import (
	"context"
	"fmt"
	"github.com/kotalco/core-api/k8s"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/logger"
	stacksv1alpha1 "github.com/kotalco/kotal/apis/stacks/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type stacksService struct{}

type IService interface {
	Get(types.NamespacedName) (stacksv1alpha1.Node, restErrors.IRestErr)
	Create(StacksDto) (stacksv1alpha1.Node, restErrors.IRestErr)
	List(namespace string) (stacksv1alpha1.NodeList, restErrors.IRestErr)
	Count(namespace string) (int, restErrors.IRestErr)
	Delete(*stacksv1alpha1.Node) restErrors.IRestErr
	Update(StacksDto, *stacksv1alpha1.Node) restErrors.IRestErr
}

var (
	k8sClient = k8s.NewClientService()
)

func NewStacksService() IService {
	return stacksService{}
}

// Create creates stacks node from spec
func (service stacksService) Create(dto StacksDto) (node stacksv1alpha1.Node, restErr restErrors.IRestErr) {
	node.ObjectMeta = dto.ObjectMetaFromMetadataDto()
	node.Spec = stacksv1alpha1.NodeSpec{
		Network:     dto.Network,
		Image:       dto.Image,
		BitcoinNode: *dto.BitcoinNode,
		RPC:         true,
	}

	k8s.DefaultResources(&node.Spec.Resources)

	if err := k8sClient.Create(context.Background(), &node); err != nil {
		if apiErrors.IsAlreadyExists(err) {
			restErr = restErrors.NewBadRequestError(fmt.Sprintf("node by name %s is already exits", node.Name))
			return
		}
		go logger.Error(service.Create, err)
		restErr = restErrors.NewInternalServerError("failed to create node")
		return
	}

	return
}

// Get returns a single stacks node by name
func (service stacksService) Get(namespacedName types.NamespacedName) (node stacksv1alpha1.Node, restErr restErrors.IRestErr) {
	if err := k8sClient.Get(context.Background(), namespacedName, &node); err != nil {
		if apiErrors.IsNotFound(err) {
			restErr = restErrors.NewNotFoundError(fmt.Sprintf("node by name %s doesn't exist", namespacedName.Name))
			return
		}
		go logger.Error(service.Get, err)
		restErr = restErrors.NewInternalServerError(fmt.Sprintf("can't get node by name %s", namespacedName.Name))
		return
	}

	return
}

// List returns all stacks nodes
func (service stacksService) List(namespace string) (list stacksv1alpha1.NodeList, restErr restErrors.IRestErr) {
	err := k8sClient.List(context.Background(), &list, client.InNamespace(namespace))
	if err != nil {
		go logger.Error(service.List, err)
		restErr = restErrors.NewInternalServerError("failed to get all nodes")
		return
	}

	return
}

// Count returns all nodes length
func (service stacksService) Count(namespace string) (count int, restErr restErrors.IRestErr) {
	nodes := &stacksv1alpha1.NodeList{}
	err := k8sClient.List(context.Background(), nodes, client.InNamespace(namespace))
	if err != nil {
		go logger.Error(service.Count, err)
		restErr = restErrors.NewInternalServerError("failed to get all nodes")
		return
	}

	return len(nodes.Items), nil
}

// Update updates a single node by name from spec
func (service stacksService) Update(dto StacksDto, node *stacksv1alpha1.Node) (restErr restErrors.IRestErr) {
	if dto.Image != "" {
		node.Spec.Image = dto.Image
	}
	if dto.RPC != nil {
		node.Spec.RPC = *dto.RPC
	}
	if dto.P2PPort != 0 {
		node.Spec.P2PPort = dto.P2PPort
	}
	if dto.RPCPort != 0 {
		node.Spec.RPCPort = dto.RPCPort
	}
	if dto.NodePrivateKeySecretName != nil {
		node.Spec.NodePrivateKeySecretName = *dto.NodePrivateKeySecretName
	}
	if dto.SeedPrivateKeySecretName != nil {
		node.Spec.SeedPrivateKeySecretName = *dto.SeedPrivateKeySecretName
	}
	if dto.Miner != nil {
		node.Spec.Miner = *dto.Miner
	}
	if dto.MineMicroBlocks != nil {
		node.Spec.MineMicroblocks = *dto.MineMicroBlocks
	}

	if dto.BitcoinNode != nil {
		if dto.BitcoinNode.Endpoint != "" {
			node.Spec.BitcoinNode.Endpoint = dto.BitcoinNode.Endpoint
		}
		if dto.BitcoinNode.RpcPort != 0 {
			node.Spec.BitcoinNode.RpcPort = dto.BitcoinNode.RpcPort
		}
		if dto.BitcoinNode.P2pPort != 0 {
			node.Spec.BitcoinNode.P2pPort = dto.BitcoinNode.P2pPort
		}
		if dto.BitcoinNode.RpcUsername != "" {
			node.Spec.BitcoinNode.RpcUsername = dto.BitcoinNode.RpcUsername
		}
		if dto.BitcoinNode.RpcPasswordSecretName != "" {
			node.Spec.BitcoinNode.RpcPasswordSecretName = dto.BitcoinNode.RpcPasswordSecretName
		}
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

// Delete deletes stacks node by name
func (service stacksService) Delete(node *stacksv1alpha1.Node) (restErr restErrors.IRestErr) {
	if err := k8sClient.Delete(context.Background(), node); err != nil {
		go logger.Error(service.Delete, err)
		restErr = restErrors.NewInternalServerError(fmt.Sprintf("can't delete node by name %s", node.Name))
		return
	}
	return
}
