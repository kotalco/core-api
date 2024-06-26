package filecoin

import (
	"context"
	"fmt"
	"github.com/kotalco/core-api/k8s"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/logger"
	filecoinv1alpha1 "github.com/kotalco/kotal/apis/filecoin/v1alpha1"
	sharedAPIs "github.com/kotalco/kotal/apis/shared"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type filecoinService struct{}

type IService interface {
	Get(types.NamespacedName) (filecoinv1alpha1.Node, restErrors.IRestErr)
	Create(FilecoinDto) (filecoinv1alpha1.Node, restErrors.IRestErr)
	Update(FilecoinDto, *filecoinv1alpha1.Node) restErrors.IRestErr
	List(namespace string) (filecoinv1alpha1.NodeList, restErrors.IRestErr)
	Delete(*filecoinv1alpha1.Node) restErrors.IRestErr
	Count(namespace string) (int, restErrors.IRestErr)
}

var (
	k8sClient = k8s.NewClientService()
)

func NewFilecoinService() IService {
	return filecoinService{}
}

// Get gets a single filecoin node by name
func (service filecoinService) Get(namespacedName types.NamespacedName) (node filecoinv1alpha1.Node, restErr restErrors.IRestErr) {
	if err := k8sClient.Get(context.Background(), namespacedName, &node); err != nil {
		if apiErrors.IsNotFound(err) {
			restErr = restErrors.NewNotFoundError(fmt.Sprintf("node by name %s doesn't exit", namespacedName.Name))
			return
		}
		go logger.Error(service.Get, err)
		restErr = restErrors.NewInternalServerError(fmt.Sprintf("can't get node by name %s", namespacedName.Name))
		return
	}

	return
}

// Create creates filecoin node from spec
func (service filecoinService) Create(dto FilecoinDto) (node filecoinv1alpha1.Node, restErr restErrors.IRestErr) {
	node.ObjectMeta = dto.ObjectMetaFromMetadataDto()
	node.Spec = filecoinv1alpha1.NodeSpec{
		Network: filecoinv1alpha1.FilecoinNetwork(dto.Network),
		Image:   dto.Image,
		API:     true,
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
			restErr = restErrors.NewBadRequestError(fmt.Sprintf("node by name %+v already exits", dto))
			return
		}
		go logger.Error(service.Create, err)
		restErr = restErrors.NewInternalServerError("failed to create node")
		return
	}

	return
}

// Update updates filecoin node by name from spec
func (service filecoinService) Update(dto FilecoinDto, node *filecoinv1alpha1.Node) (restErr restErrors.IRestErr) {
	if dto.API != nil {
		node.Spec.API = *dto.API
	}

	if dto.APIPort != 0 {
		node.Spec.APIPort = dto.APIPort
	}

	if dto.APIRequestTimeout != 0 {
		node.Spec.APIRequestTimeout = dto.APIRequestTimeout
	}

	if dto.DisableMetadataLog != nil {
		node.Spec.DisableMetadataLog = *dto.DisableMetadataLog
	}

	if dto.P2PPort != 0 {
		node.Spec.P2PPort = dto.P2PPort
	}

	if dto.IPFSPeerEndpoint != nil {
		node.Spec.IPFSPeerEndpoint = *dto.IPFSPeerEndpoint
	}

	if dto.IPFSOnlineMode != nil {
		node.Spec.IPFSOnlineMode = *dto.IPFSOnlineMode
	}

	if dto.IPFSForRetrieval != nil {
		node.Spec.IPFSForRetrieval = *dto.IPFSForRetrieval
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

// List returns all filecoin nodes
func (service filecoinService) List(namespace string) (list filecoinv1alpha1.NodeList, restErr restErrors.IRestErr) {
	if err := k8sClient.List(context.Background(), &list, client.InNamespace(namespace)); err != nil {
		go logger.Error(service.List, err)
		restErr = restErrors.NewInternalServerError("failed to get all nodes")
		return
	}
	return
}

// Count returns total number of filecoin nodes
func (service filecoinService) Count(namespace string) (count int, restErr restErrors.IRestErr) {
	nodes := &filecoinv1alpha1.NodeList{}
	if err := k8sClient.List(context.Background(), nodes, client.InNamespace(namespace)); err != nil {
		go logger.Error(service.Count, err)
		restErr = restErrors.NewInternalServerError("failed to count filecoin nodes")
		return
	}

	return len(nodes.Items), nil
}

// Delete deletes ethereum 2.0 filecoin node by name
func (service filecoinService) Delete(node *filecoinv1alpha1.Node) (restErr restErrors.IRestErr) {
	if err := k8sClient.Delete(context.Background(), node); err != nil {
		go logger.Error(service.Delete, err)
		restErr = restErrors.NewInternalServerError(fmt.Sprintf("can't delte node by name %s", node.Name))
		return
	}
	return nil
}
