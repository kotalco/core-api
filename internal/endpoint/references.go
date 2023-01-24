package endpoint

var references = map[protocol]map[kind][]string{
	ipfs: {
		peer: {
			"google.com",
			"kotal.co",
		},
		clusterpeer: {
			"google.com",
			"kotal.co",
		},
	},
	ethereum: {
		node: {
			"google.com",
			"kotal.co",
		},
	},
	ethereum2: {
		beaconnode: {
			"google.com",
			"kotal.co",
		},
		validator: {
			"google.com",
			"kotal.co",
		},
	},
}
