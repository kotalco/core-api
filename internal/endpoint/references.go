package endpoint

var references = map[protocol]map[kind]map[port][]string{
	ipfs: {
		peer: {
			api:     []string{"google.com", "kotal.co"},
			gateway: []string{"google.com", "kotal.co"},
		},
	},
	ethereum: {
		node: {
			rpc:     []string{"google.com", "kotal.co"},
			ws:      []string{"google.com", "kotal.co"},
			qraphql: []string{"google.com", "kotal.co"},
		},
	},
	ethereum2: {
		beaconnode: {
			rpc: []string{"google.com", "kotal.co"},
		},
		validator: {
			ws: []string{"google.com", "kotal.co"},
		},
	},
	polkadot: {
		node: {
			rpc: []string{"google.com", "kotal.co"},
			ws:  []string{"google.com", "kotal.co"},
		},
	},
}
