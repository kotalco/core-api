package endpoint

var endpointExamples = map[string]map[string]string{
	"ipfs-peer":        {"api": "curl -X POST ${route}/api/v0/swarm/peers", "gateway": "curl ${route}/ipfs/QmQPeNsJPyVWPFDVHb77w8G42Fvo15z4bG2X8D2GhfbSXc/readme"},
	"ipfs-clusterpeer": {"api": "curl -X POST ${route}/api/v0/swarm/peers", "gateway": "curl ${route}/ipfs/QmQPeNsJPyVWPFDVHb77w8G42Fvo15z4bG2X8D2GhfbSXc/readme"},
	"ethereum-node": {
		"rpc": `curl -X POST -H 'content-type: application/json' --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' ${route}`,
		"ws": `// npm install web3

const Web3 = require('web3');
const web3 = new Web3("${route}");

web3.eth.getBlockNumber().then(function (blockNumber) {
    console.log('block number is ${blockNumber}')
    process.exit(0)
})`,
	},
}
