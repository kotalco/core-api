package ipfs_peer

import (
	"github.com/kotalco/core-api/k8s"
	"github.com/kotalco/core-api/pkg/time"
	ipfsv1alpha1 "github.com/kotalco/kotal/apis/ipfs/v1alpha1"
	sharedAPI "github.com/kotalco/kotal/apis/shared"
)

// Peer is IPFS peer
// TODO: update with SwarmKeySecret and Resources
type PeerDto struct {
	time.Time
	k8s.MetaDataDto
	InitProfiles []string `json:"initProfiles"`
	APIPort      uint     `json:"apiPort"`
	GatewayPort  uint     `json:"gatewayPort"`
	Routing      string   `json:"routing"`
	Profiles     []string `json:"profiles"`
	API          *bool    `json:"api"`
	Gateway      *bool    `json:"gateway"`
	Image        string   `json:"image"`
	sharedAPI.Resources
}

type PeerListDto []PeerDto

// FromIPFSPeer creates peer model from IPFS peer
func (dto PeerDto) FromIPFSPeer(peer ipfsv1alpha1.Peer) PeerDto {
	var profiles, initProfiles []string

	// init profiles
	for _, profile := range peer.Spec.InitProfiles {
		initProfiles = append(initProfiles, string(profile))
	}

	// profiles
	for _, profile := range peer.Spec.Profiles {
		profiles = append(profiles, string(profile))
	}

	dto.Name = peer.Name
	dto.Time = time.Time{CreatedAt: peer.CreationTimestamp.UTC().Format(time.JavascriptISOString)}
	dto.APIPort = peer.Spec.APIPort
	dto.GatewayPort = peer.Spec.GatewayPort
	dto.Routing = string(peer.Spec.Routing)
	dto.Profiles = profiles
	dto.InitProfiles = initProfiles
	dto.CPU = peer.Spec.CPU
	dto.CPULimit = peer.Spec.CPULimit
	dto.Memory = peer.Spec.Memory
	dto.MemoryLimit = peer.Spec.MemoryLimit
	dto.Storage = peer.Spec.Storage
	dto.StorageClass = peer.Spec.StorageClass
	dto.API = &peer.Spec.API
	dto.Gateway = &peer.Spec.Gateway
	dto.Image = peer.Spec.Image

	return dto
}

func (peerListDto PeerListDto) FromIPFSPeer(peers []ipfsv1alpha1.Peer) PeerListDto {
	result := make(PeerListDto, len(peers))
	for index, v := range peers {
		result[index] = PeerDto{}.FromIPFSPeer(v)
	}
	return result
}
