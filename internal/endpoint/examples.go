package endpoint

type (
	protocol string
	kind     string
	port     string
)

const (
	chainlink protocol = "chainlink"
	ethereum  protocol = "ethereum"
	ethereum2 protocol = "ethereum2"
	filecoin  protocol = "filecoin"
	ipfs      protocol = "ipfs"
	near      protocol = "near"
	polkadot  protocol = "polkadot"
)

const (
	beaconnode  kind = "beaconnode"
	clusterpeer kind = "clusterpeer"
	node        kind = "node"
	peer        kind = "peer"
	validator   kind = "validator"
)
const (
	api     port = "api"
	gateway port = "gateway"
	grpc    port = "grpc"
	qraphql port = "graphql"
	rest    port = "rest"
	rpc     port = "rpc"
	ws      port = "ws"
)

var examples = map[protocol]map[kind]map[port]string{
	ipfs: {
		peer: {
			api:     `curl -X POST ${route}/api/v0/swarm/peers`,
			gateway: `curl ${route}/ipfs/QmQPeNsJPyVWPFDVHb77w8G42Fvo15z4bG2X8D2GhfbSXc/readme`,
		},
		clusterpeer: {
			api:  "curl ${route}/pins",
			rest: "curl ${route}/id",
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
			rest: `curl ${route}/eth/v1/beacon/genesis`,
			rpc:  `curl ${route}/eth/v1/beacon/genesis`,
			grpc: `curl ${route}/eth/v1/beacon/genesis`,
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
	chainlink: {
		node: {
			api: `# replace with your api credentials
echo "your@email.com\nyourpassword" > credentials
NODE_ENDPOINT=${route}
# login
chainlink --remote-node-url $NODE_ENDPOINT admin login --file credentials --bypass-version-check
# list ethereum keys
chainlink --remote-node-url $NODE_ENDPOINT keys eth list
`,
		},
	},
	near: {
		node: {
			rpc: `curl -X POST -H 'content-type: application/json' --data '{"jsonrpc":"2.0","method":"block", "params": {"finality": "final"}, "id":1}' ${route}`,
		},
	},
	filecoin: {
		node: {
			rpc: `curl -X POST -H 'Content-Type: application/json' --data '{"jsonrpc":"2.0","id":1,"method":"Filecoin.ChainHead","params":[]}' ${route}`,
		},
	},
}
