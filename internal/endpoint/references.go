package endpoint

var references = map[protocol]map[kind]map[port][]string{
	ipfs: {
		peer: {
			api: {
				"https://docs.ipfs.tech/reference/kubo/rpc/",
			},
			gateway: {
				"https://docs.ipfs.tech/concepts/ipfs-gateway/",
			},
		},
		clusterpeer: {
			api: {
				"https://ipfscluster.io/documentation/reference/pinsvc_api/",
				"https://ipfs.github.io/pinning-services-api-spec/",
			},
			rest: {
				"https://ipfscluster.io/documentation/reference/api/",
			},
		},
	},
	ethereum: {
		node: {
			rpc: {
				"https://ethereum.org/en/developers/docs/apis/json-rpc/",
				"https://ethereum.github.io/execution-apis/api-documentation/",
				"https://besu.hyperledger.org/en/stable/public-networks/reference/api/",
				"https://geth.ethereum.org/docs/interacting-with-geth/rpc",
			},
			ws: {
				"https://ethereum.org/en/developers/tutorials/using-websockets/",
				"https://besu.hyperledger.org/en/stable/public-networks/how-to/use-besu-api/json-rpc/#websocket",
			},
			qraphql: {
				"https://besu.hyperledger.org/en/stable/public-networks/how-to/use-besu-api/graphql/",
				"https://ethereum.org/en/developers/tutorials/the-graph-fixing-web3-data-querying/",
			},
		},
	},
	ethereum2: {
		beaconnode: {
			rest: {
				"https://ethereum.github.io/beacon-APIs/",
			},
			rpc: {
				"https://ethereum.github.io/beacon-APIs/",
			},
			grpc: {
				"https://ethereum.github.io/beacon-APIs/",
			},
		},
	},
	polkadot: {
		node: {
			rpc: {
				"https://polkadot.js.org/docs/substrate/rpc/",
			},
			ws: {
				"https://polkadot.js.org/docs/api/start/create/",
				"https://wiki.polkadot.network/docs/maintain-wss",
			},
		},
	},
	chainlink: {
		node: {
			api: {
				"https://github.com/smartcontractkit/chainlink#build-chainlink",
			},
		},
	},
	near: {
		node: {
			rpc: {
				"https://docs.near.org/api/rpc/introduction",
				"https://docs.near.org/tools/near-api-js/quick-reference",
			},
		},
	},
	filecoin: {
		node: {
			api: {
				"https://docs.filecoin.io/developers/reference/json-rpc/introduction/",
			},
		},
	},
}
