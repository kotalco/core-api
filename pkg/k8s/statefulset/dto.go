package statefulset

var StatFulSetProtocolList = struct {
	Bitcoin   string
	Chainlink string
	Ethereum  string
	Ethereum2 string
	Filecoin  string
	Ipfs      string
	Near      string
	Polkadot  string
	Stacks    string
}{
	Bitcoin:   "bitcoin",
	Chainlink: "chainlink",
	Ethereum:  "ethereum",
	Ethereum2: "ethereum2",
	Filecoin:  "filecoin",
	Ipfs:      "ipfs",
	Near:      "near",
	Polkadot:  "polkadot",
	Stacks:    "stacks",
}

type CountResponseDto struct {
	Bitcoin   int `json:"bitcoin"`
	Chainlink int `json:"chainlink"`
	Ethereum  int `json:"ethereum"`
	Ethereum2 int `json:"ethereum_2"`
	Filecoin  int `json:"filecoin"`
	Ipfs      int `json:"ipfs"`
	Near      int `json:"near"`
	Polkadot  int `json:"polkadot"`
	Stacks    int `json:"stacks"`
}
