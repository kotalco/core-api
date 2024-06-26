package ipfs_cluster_peer

import (
	"context"
	"fmt"
	"github.com/kotalco/core-api/k8s"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/logger"
	ipfsv1alpha1 "github.com/kotalco/kotal/apis/ipfs/v1alpha1"
	sharedAPIs "github.com/kotalco/kotal/apis/shared"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ipfsClusterPeerService struct{}

type IService interface {
	Get(name types.NamespacedName) (ipfsv1alpha1.ClusterPeer, restErrors.IRestErr)
	Create(ClusterPeerDto) (ipfsv1alpha1.ClusterPeer, restErrors.IRestErr)
	Update(ClusterPeerDto, *ipfsv1alpha1.ClusterPeer) restErrors.IRestErr
	List(namespace string) (ipfsv1alpha1.ClusterPeerList, restErrors.IRestErr)
	Delete(*ipfsv1alpha1.ClusterPeer) restErrors.IRestErr
	Count(namespace string) (int, restErrors.IRestErr)
}

var (
	k8sClient = k8s.NewClientService()
)

func NewIpfsClusterPeerService() IService {
	return ipfsClusterPeerService{}
}

// Get gets a single IPFS peer by name
func (service ipfsClusterPeerService) Get(namespacedName types.NamespacedName) (peer ipfsv1alpha1.ClusterPeer, restErr restErrors.IRestErr) {
	if err := k8sClient.Get(context.Background(), namespacedName, &peer); err != nil {
		if apiErrors.IsNotFound(err) {
			restErr = restErrors.NewNotFoundError(fmt.Sprintf("cluster peer by name %s doesn't exit", namespacedName.Name))
			return
		}
		go logger.Error(service.Get, err)
		restErr = restErrors.NewInternalServerError(fmt.Sprintf("can't get cluster peer by name %s", peer.Name))
		return
	}

	return
}

// Create creates IPFS peer from spec
func (service ipfsClusterPeerService) Create(dto ClusterPeerDto) (peer ipfsv1alpha1.ClusterPeer, restErr restErrors.IRestErr) {
	peer.ObjectMeta = dto.ObjectMetaFromMetadataDto()
	peer.Spec = ipfsv1alpha1.ClusterPeerSpec{
		Image: dto.Image,
		Resources: sharedAPIs.Resources{
			StorageClass: dto.StorageClass,
		},
	}

	k8s.DefaultResources(&peer.Spec.Resources)

	if dto.PeerEndpoint != "" {
		peer.Spec.PeerEndpoint = dto.PeerEndpoint
	}

	if dto.Consensus != "" {
		peer.Spec.Consensus = ipfsv1alpha1.ConsensusAlgorithm(dto.Consensus)
	}

	if dto.ID != "" {
		peer.Spec.ID = dto.ID
	}

	if dto.PrivatekeySecretName != "" {
		peer.Spec.PrivateKeySecretName = dto.PrivatekeySecretName
	}

	if len(dto.TrustedPeers) != 0 {
		peer.Spec.TrustedPeers = dto.TrustedPeers
	}

	if len(dto.BootstrapPeers) != 0 {
		peer.Spec.BootstrapPeers = dto.BootstrapPeers
	}

	if dto.ClusterSecretName != "" {
		peer.Spec.ClusterSecretName = dto.ClusterSecretName
	}

	if dto.CPU != "" {
		peer.Spec.CPU = dto.CPU
	}
	if dto.CPULimit != "" {
		peer.Spec.CPULimit = dto.CPULimit
	}
	if dto.Memory != "" {
		peer.Spec.Memory = dto.Memory
	}
	if dto.MemoryLimit != "" {
		peer.Spec.MemoryLimit = dto.MemoryLimit
	}
	if dto.Storage != "" {
		peer.Spec.Storage = dto.Storage
	}

	if os.Getenv("MOCK") == "true" {
		peer.Default()
	}

	if err := k8sClient.Create(context.Background(), &peer); err != nil {
		if apiErrors.IsAlreadyExists(err) {
			restErr = restErrors.NewBadRequestError(fmt.Sprintf("cluster peer by name %s already exits", peer.Name))
			return
		}
		go logger.Error(service.Create, err)
		restErr = restErrors.NewInternalServerError("failed to create cluster peer")
		return
	}

	return
}

// Update updates IPFS peer by name from spec
func (service ipfsClusterPeerService) Update(dto ClusterPeerDto, peer *ipfsv1alpha1.ClusterPeer) (restErr restErrors.IRestErr) {
	if dto.PeerEndpoint != "" {
		peer.Spec.PeerEndpoint = dto.PeerEndpoint
	}

	if len(dto.BootstrapPeers) != 0 {
		peer.Spec.BootstrapPeers = dto.BootstrapPeers
	}

	if dto.CPU != "" {
		peer.Spec.CPU = dto.CPU
	}
	if dto.CPULimit != "" {
		peer.Spec.CPULimit = dto.CPULimit
	}
	if dto.Memory != "" {
		peer.Spec.Memory = dto.Memory
	}
	if dto.MemoryLimit != "" {
		peer.Spec.MemoryLimit = dto.MemoryLimit
	}
	if dto.Storage != "" {
		peer.Spec.Storage = dto.Storage
	}
	if dto.Image != "" {
		peer.Spec.Image = dto.Image
	}

	if os.Getenv("MOCK") == "true" {
		peer.Default()
	}

	pod := &corev1.Pod{}
	podIsPending := false
	if dto.CPU != "" || dto.Memory != "" {
		key := types.NamespacedName{
			Namespace: peer.Namespace,
			Name:      fmt.Sprintf("%s-0", peer.Name),
		}
		err := k8sClient.Get(context.Background(), key, pod)
		if apiErrors.IsNotFound(err) {
			go logger.Error(service.Update, err)
			restErr = restErrors.NewBadRequestError(fmt.Sprintf("pod by name %s doesn't exit", key.Name))
			return
		}
		podIsPending = pod.Status.Phase == corev1.PodPending
	}

	if err := k8sClient.Update(context.Background(), peer); err != nil {
		go logger.Error(service.Update, err)
		restErr = restErrors.NewInternalServerError(fmt.Sprintf("can't update cluster peer by name %s", peer.Name))
		return
	}

	if podIsPending {
		err := k8sClient.Delete(context.Background(), pod)
		if err != nil {
			go logger.Error(service.Update, err)
			restErr = restErrors.NewInternalServerError(fmt.Sprintf("can't update cluster peer by name %s", peer.Name))
			return
		}
	}

	return
}

// List returns all IPFS peers
func (service ipfsClusterPeerService) List(namespace string) (list ipfsv1alpha1.ClusterPeerList, restErr restErrors.IRestErr) {
	if err := k8sClient.List(context.Background(), &list, client.InNamespace(namespace)); err != nil {
		go logger.Error(service.List, err)
		restErr = restErrors.NewInternalServerError("failed to get all peers")
		return
	}

	return
}

// Count returns total number of IPFS peers
func (service ipfsClusterPeerService) Count(namespace string) (count int, restErr restErrors.IRestErr) {
	peers := &ipfsv1alpha1.ClusterPeerList{}
	if err := k8sClient.List(context.Background(), peers, client.InNamespace(namespace)); err != nil {
		go logger.Error(service.Count, err)
		restErr = restErrors.NewInternalServerError("failed to count all cluster peers")
		return
	}

	return len(peers.Items), nil
}

// Delete deletes ethereum 2.0 IPFS peer by name
func (service ipfsClusterPeerService) Delete(peer *ipfsv1alpha1.ClusterPeer) (restErr restErrors.IRestErr) {
	if err := k8sClient.Delete(context.Background(), peer); err != nil {
		go logger.Error(service.Delete, err)
		restErr = restErrors.NewInternalServerError(fmt.Sprintf("can't delete cluster peer by name %s", peer.Name))
		return
	}

	return
}
