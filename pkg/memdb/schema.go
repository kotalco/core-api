package memdb

import (
	"github.com/hashicorp/go-memdb"
)

type Student struct {
	Id    int
	Name  string
	Score int
}

// Define the DB schema
var schema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		"endpointExample": &memdb.TableSchema{
			Name: "endpointExample",
			Indexes: map[string]*memdb.IndexSchema{
				"id": &memdb.IndexSchema{
					Name:    "id",
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "Id"},
				},
				"kind": &memdb.IndexSchema{ // represents the kind of this endpoint eg, ipfs-peer
					Name:    "kind",
					Unique:  false,
					Indexer: &memdb.StringFieldIndex{Field: "Kind"},
				},
				"name": &memdb.IndexSchema{ // represents the port in which the user goning to access this endpoint throw eg, api,rpc,wss
					Name:    "name",
					Unique:  false,
					Indexer: &memdb.StringFieldIndex{Field: "Name"},
				},
				"example": &memdb.IndexSchema{ // the example of this specified endpoint which give the user a hind how to use the endpoint
					Name:    "example",
					Unique:  false,
					Indexer: &memdb.StringFieldIndex{Field: "Example"},
				},
			},
		},
		"endpointRef": &memdb.TableSchema{
			Name: "endpointRef",
			Indexes: map[string]*memdb.IndexSchema{
				"id": &memdb.IndexSchema{
					Name:    "id",
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "Id"},
				},
				"kind": &memdb.IndexSchema{ // represents the kind of this endpoint eg, ipfs-peer
					Name:    "kind",
					Unique:  false,
					Indexer: &memdb.StringFieldIndex{Field: "Kind"},
				},
				"ref": &memdb.IndexSchema{ // a slice of references for the specified protocol kind
					Name:    "ref",
					Unique:  false,
					Indexer: &memdb.StringSliceFieldIndex{Field: "Ref"},
				},
			},
		},
	},
}
