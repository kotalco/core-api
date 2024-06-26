package filecoin

import (
	"github.com/kotalco/core-api/k8s"
	"github.com/kotalco/core-api/pkg/time"
	filecoinv1alpha1 "github.com/kotalco/kotal/apis/filecoin/v1alpha1"
	sharedAPI "github.com/kotalco/kotal/apis/shared"
)

// Node is Filecoin node
type FilecoinDto struct {
	time.Time
	k8s.MetaDataDto
	Network            string  `json:"network"`
	API                *bool   `json:"api"`
	APIPort            uint    `json:"apiPort"`
	APIRequestTimeout  uint    `json:"apiRequestTimeout"`
	DisableMetadataLog *bool   `json:"disableMetadataLog"`
	P2PPort            uint    `json:"p2pPort"`
	IPFSPeerEndpoint   *string `json:"ipfsPeerEndpoint"`
	IPFSOnlineMode     *bool   `json:"ipfsOnlineMode"`
	IPFSForRetrieval   *bool   `json:"ipfsForRetrieval"`
	Image              string  `json:"image"`
	sharedAPI.Resources
}

type FilecoinListDto []FilecoinDto

// FromFilecoinNode creates node dto from Filecoin node
func (dto FilecoinDto) FromFilecoinNode(node filecoinv1alpha1.Node) FilecoinDto {

	dto.Name = node.Name
	dto.Time = time.Time{CreatedAt: node.CreationTimestamp.UTC().Format(time.JavascriptISOString)}
	dto.Network = string(node.Spec.Network)
	dto.API = &node.Spec.API
	dto.APIPort = node.Spec.APIPort
	dto.APIRequestTimeout = node.Spec.APIRequestTimeout
	dto.DisableMetadataLog = &node.Spec.DisableMetadataLog
	dto.P2PPort = node.Spec.P2PPort
	dto.IPFSPeerEndpoint = &node.Spec.IPFSPeerEndpoint
	dto.IPFSOnlineMode = &node.Spec.IPFSOnlineMode
	dto.IPFSForRetrieval = &node.Spec.IPFSForRetrieval
	dto.CPU = node.Spec.CPU
	dto.CPULimit = node.Spec.CPULimit
	dto.Memory = node.Spec.Memory
	dto.MemoryLimit = node.Spec.MemoryLimit
	dto.Storage = node.Spec.Storage
	dto.StorageClass = node.Spec.StorageClass
	dto.Image = node.Spec.Image

	return dto
}

// FromFilecoinNode creates node dto from Filecoin node list
func (filecoinListDto FilecoinListDto) FromFilecoinNode(nodes []filecoinv1alpha1.Node) FilecoinListDto {
	result := make(FilecoinListDto, len(nodes))
	for index, v := range nodes {
		result[index] = FilecoinDto{}.FromFilecoinNode(v)
	}
	return result
}
