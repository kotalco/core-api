package endpoint

type (
	protocol string
	kind     string
	port     string
)

const (
	ipfs      protocol = "ipfs"
	ethereum  protocol = "ethereum"
	ethereum2 protocol = "ethereum2"
	polkadot  protocol = "polkadot"
)
const (
	peer        kind = "peer"
	clusterpeer kind = "clusterpeer"
	node        kind = "node"
	beaconnode  kind = "beaconnode"
	validator   kind = "validator"
)
const (
	api     port = "api"
	gateway port = "gateway"
	rpc     port = "rpc"
	ws      port = "ws"
	qraphql port = "graphql"
)

var examples = map[protocol]map[kind]map[port]string{
	ipfs: {
		peer: {
			api:     `curl -X POST ${route}/api/v0/swarm/peers`,
			gateway: `curl ${route}/ipfs/QmQPeNsJPyVWPFDVHb77w8G42Fvo15z4bG2X8D2GhfbSXc/readme`,
		},
	},
	ethereum: {
		node: {
			rpc: `curl -X POST -H 'content-type: application/json' --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' ${route}`,
			ws: `// npm install web3
            const Web3 = require('web3');
            const web3 = new Web3("${route}");
            web3.eth.getBlockNumber().then(function (blockNumber) {
            console.log('block number is ${blockNumber}')
            process.exit(0)})`,
			qraphql: `curl -X POST -H "Content-Type: application/json" --data '{ "query": "{syncing{startingBlock currentBlock highestBlock}}"}' ${route}`,
		},
	},
	ethereum2: {
		beaconnode: {
			rpc: `curl -X POST -H 'content-type: application/json' --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' ${route}`,
		},
		validator: {
			ws: `// npm install web3
           const Web3 = require('web3');
           const web3 = new Web3("${route}");
           web3.eth.getBlockNumber().then(function (blockNumber) {
           console.log('block number is ${blockNumber}')
           process.exit(0)})`,
		},
	},
	polkadot: {
		node: {
			rpc: `curl -X POST -H 'content-type: application/json' --data '{"jsonrpc":"2.0","method":"system_chain","params":[],"id":1}' ${route}`,
			ws: `// npm install @polkadot/api
      				import { ApiPromise, WsProvider } from '@polkadot/api';
					import "os"

					const provider = new WsProvider('${route}');
					const api = await ApiPromise.create({ provider: provider });
					console.log("Runtime chain is " + api.runtimeChain.toString());`,
		},
	},
}
