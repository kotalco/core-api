package memdb

type EndpointExample struct {
	Id      string
	Kind    string
	Name    string
	Example string
}

type EndpointRef struct {
	Id   string
	Kind string
	Ref  []string
}
