package svc

type SvcDto struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
}

var AvailableProtocol = func(protocol string) bool {
	switch protocol {
	case "ws", "p2p":
		return false
	default:
		return true
	}
}
