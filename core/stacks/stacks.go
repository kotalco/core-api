package stacks

import (
	"github.com/kotalco/core-api/k8s"
	"github.com/kotalco/core-api/pkg/time"
	sharedAPI "github.com/kotalco/kotal/apis/shared"
	stacksv1alpha1 "github.com/kotalco/kotal/apis/stacks/v1alpha1"
)

type StacksDto struct {
	time.Time
	k8s.MetaDataDto
	Image                    string                       `json:"image"`
	Network                  stacksv1alpha1.StacksNetwork `json:"network"`
	RPC                      *bool                        `json:"rpc"`
	P2PPort                  uint                         `json:"p2pPort"`
	RPCPort                  uint                         `json:"rpcPort"`
	NodePrivateKeySecretName *string                      `json:"nodePrivateKeySecretName"`
	SeedPrivateKeySecretName *string                      `json:"seedPrivateKeySecretName"`
	Miner                    *bool                        `json:"miner"`
	MineMicroBlocks          *bool                        `json:"mineMicroBlocks"`
	BitcoinNode              *stacksv1alpha1.BitcoinNode  `json:"bitcoinNode"`
	sharedAPI.Resources
}

type StacksListDto []StacksDto

func (dto StacksDto) FromStacksNode(n stacksv1alpha1.Node) StacksDto {
	dto.Name = n.Name
	dto.Time = time.Time{CreatedAt: n.CreationTimestamp.UTC().Format(time.JavascriptISOString)}
	dto.Image = n.Spec.Image
	dto.Network = n.Spec.Network
	dto.RPC = &n.Spec.RPC
	dto.P2PPort = n.Spec.P2PPort
	dto.RPCPort = n.Spec.RPCPort
	dto.NodePrivateKeySecretName = &n.Spec.NodePrivateKeySecretName
	dto.SeedPrivateKeySecretName = &n.Spec.SeedPrivateKeySecretName
	dto.Miner = &n.Spec.Miner
	dto.MineMicroBlocks = &n.Spec.MineMicroblocks
	dto.BitcoinNode = &n.Spec.BitcoinNode
	dto.CPU = n.Spec.CPU
	dto.CPULimit = n.Spec.CPULimit
	dto.Memory = n.Spec.Memory
	dto.MemoryLimit = n.Spec.MemoryLimit
	dto.Storage = n.Spec.Storage
	dto.StorageClass = n.Spec.StorageClass
	return dto
}

func (nodes StacksListDto) FromStacksNode(models []stacksv1alpha1.Node) StacksListDto {
	result := make(StacksListDto, len(models))
	for index, model := range models {
		result[index] = StacksDto{}.FromStacksNode(model)
	}
	return result
}
