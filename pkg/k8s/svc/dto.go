package svc

type SvcDto struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
}

var AvailableProtocol = func(protocol string) bool {
	switch protocol {
	case "p2p", "metrics", "discovery", "tls", "tracing", "swarm", "swarm-udp", "prometheus":
		return false
	default:
		return true //rpc,ws,api,graphql,gateway,grpc,rest
	}
}
