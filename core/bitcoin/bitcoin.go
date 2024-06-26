package bitcoin

import (
	"github.com/kotalco/core-api/k8s"
	"github.com/kotalco/core-api/pkg/time"
	bitcoinv1alpha1 "github.com/kotalco/kotal/apis/bitcoin/v1alpha1"
	sharedAPI "github.com/kotalco/kotal/apis/shared"
)

const (
	BitcoinJsonRpcDefaultUserName           = "kotal"
	BitcoinJsonRpcDefaultUserPasswordName   = "kotal-rpc-user-password"
	BitcoinJsonRpcDefaultUserPasswordSecret = "2bbf1fdeff5f5c2dbae910a7a43776ab1d829446d7bd911c4812f7eb47f224aa"
)

type RPCUser struct {
	Username           string `json:"username"`
	PasswordSecretName string `json:"passwordSecretName"`
}

type BitcoinDto struct {
	time.Time
	k8s.MetaDataDto
	Image            string                         `json:"image"`
	Network          bitcoinv1alpha1.BitcoinNetwork `json:"network"`
	P2PPort          uint                           `json:"p2pPort"`
	RPC              *bool                          `json:"rpc"`
	RPCPort          uint                           `json:"rpcPort"`
	RPCUsers         []RPCUser                      `json:"rpcUsers"`
	Wallet           *bool                          `json:"wallet"`
	TransactionIndex *bool                          `json:"txIndex"`
	sharedAPI.Resources
}

type BitcoinListDto []BitcoinDto

func (dto BitcoinDto) FromBitcoinNode(n bitcoinv1alpha1.Node) BitcoinDto {
	dto.Name = n.Name
	dto.Time = time.Time{CreatedAt: n.CreationTimestamp.UTC().Format(time.JavascriptISOString)}
	dto.Image = n.Spec.Image
	dto.Network = n.Spec.Network
	dto.P2PPort = n.Spec.P2PPort
	dto.RPC = &n.Spec.RPC
	dto.RPCPort = n.Spec.RPCPort
	dto.RPCUsers = make([]RPCUser, 0)
	for _, v := range n.Spec.RPCUsers {
		dto.RPCUsers = append(dto.RPCUsers, RPCUser{
			Username:           v.Username,
			PasswordSecretName: v.PasswordSecretName,
		})
	}
	dto.Wallet = &n.Spec.Wallet
	dto.TransactionIndex = &n.Spec.TransactionIndex
	dto.CPU = n.Spec.CPU
	dto.CPULimit = n.Spec.CPULimit
	dto.Memory = n.Spec.Memory
	dto.MemoryLimit = n.Spec.MemoryLimit
	dto.Storage = n.Spec.Storage
	dto.StorageClass = n.Spec.StorageClass
	return dto
}

func (nodes BitcoinListDto) FromBitcoinNode(models []bitcoinv1alpha1.Node) BitcoinListDto {
	result := make(BitcoinListDto, len(models))
	for index, model := range models {
		result[index] = BitcoinDto{}.FromBitcoinNode(model)
	}
	return result
}
