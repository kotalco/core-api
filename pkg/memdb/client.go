package memdb

import (
	"github.com/hashicorp/go-memdb"
	"github.com/kotalco/community-api/pkg/logger"
	"sync"
)

var (
	memDbConnection *memdb.MemDB
	memdbOnce       sync.Once
	err             error
)

type IMemDB interface {
	Begin(memdb *memdb.MemDB, write bool)
	Commit()
	Abort()
	First(table, index string, args ...interface{}) (interface{}, error)
	Get(table, index string, args ...interface{}) (memdb.ResultIterator, error)
	Insert(table string, obj interface{}) error
}

type MemTxn struct {
	Txn *memdb.Txn
}

func NewMemDb() IMemDB {
	return &MemTxn{}
}

func OpenMemDbConnection() *memdb.MemDB {
	memdbOnce.Do(func() {
		memDbConnection, err = memdb.NewMemDB(schema)
		if err != nil {
			go logger.Warn("MEME_DB_CONNECTION", err)
		}
	})
	return memDbConnection
}

func (m *MemTxn) Begin(memdb *memdb.MemDB, write bool) {
	m.Txn = memdb.Txn(write)
}

func (m *MemTxn) Commit() {
	m.Txn.Commit()
}

func (m *MemTxn) Abort() {
	m.Txn.Abort()
}

func (m *MemTxn) First(table, index string, args ...interface{}) (interface{}, error) {
	return m.Txn.First(table, index, args...)
}
func (m *MemTxn) Get(table, index string, args ...interface{}) (memdb.ResultIterator, error) {
	return m.Txn.Get(table, index, args...)
}
func (m *MemTxn) Insert(table string, obj interface{}) error {
	return m.Txn.Insert(table, obj)
}
