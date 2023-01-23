package memdb

import (
	"github.com/google/uuid"
	"github.com/kotalco/community-api/pkg/logger"
)

func SeedMemDB() {
	// Insert endpointExamples
	examples := []*EndpointExample{
		&EndpointExample{Id: uuid.NewString(), Kind: "ipfs-peer", Name: "api", Example: "curl -X POST ${route}/api/v0/swarm/peers"},
		&EndpointExample{Id: uuid.NewString(), Kind: "ipfs-peer", Name: "gateway", Example: "curl -X POST ${route}/api/v0/swarm/peers"},
		&EndpointExample{Id: uuid.NewString(), Kind: "ethereum-node", Name: "rpc", Example: `curl -X POST -H 'content-type: application/json' --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' ${route}`},
		&EndpointExample{
			Id:   uuid.NewString(),
			Kind: "ethereum-node",
			Name: "ws",
			Example: `// npm install web3
                     const Web3 = require('web3');
                     const web3 = new Web3("${route}");
                     web3.eth.getBlockNumber().then(function (blockNumber) {
                     console.log('block number is ${blockNumber}')
                     process.exit(0)
                     })`,
		},
	}
	refExamples := []*EndpointRef{
		&EndpointRef{Id: uuid.NewString(), Kind: "ipfs-peer", Ref: []string{"google.com", "kotal.co"}},
		&EndpointRef{Id: uuid.NewString(), Kind: "ethereum-node", Ref: []string{"google.com", "kotal.co"}},
	}

	memdb := NewMemDb()
	memdb.Begin(OpenMemDbConnection(), true)

	for _, k := range examples {
		if err := memdb.Insert("endpointExample", k); err != nil {
			go logger.Warn("SEED_ENDPOINT_EXAMPLE", err)
		}
	}
	for _, k := range refExamples {
		if err := memdb.Insert("endpointRef", k); err != nil {
			go logger.Warn("SEED_ENDPOINT_REF_EXAMPLE", err)
		}
	}

	memdb.Commit()
}
