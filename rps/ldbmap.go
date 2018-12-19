package rps

import (
	"path/filepath"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
)

// LDBMap structure.
type LDBMap struct {
	data  map[string]string
	lock  *sync.RWMutex
	name  string
	db    *leveldb.DB
	batch *leveldb.Batch
}

// NewLDBMap creates an object of NewLDBMap structure.
func NewLDBMap(name string, dbPath string) LDBMap {
	lock := sync.RWMutex{}
	data := map[string]string{}
	db, err := leveldb.OpenFile(filepath.Join(dbPath, name), nil)
	if err != nil {
		Error.Printf("Can't get access to database:\n\tDBPath: %s",
			filepath.Join(dbPath, name))
		panic(err)
	}

	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		k, v := iter.Key(), iter.Value()
		sk, sv := string(k), string(v)
		data[sk] = sv
	}

	batch := new(leveldb.Batch)
	ldbMap := LDBMap{data, &lock, name, db, batch}

	return ldbMap
}

// Get performs non-blocking get of an object from the vault.
func (m LDBMap) Get(key string) string {
	(*m.lock).RLock()
	defer (*m.lock).RUnlock()
	return m.data[key]
}

// Put performs non-blocking put of an object into the vault.
func (m LDBMap) Put(key string, value string) error {
	(*m.lock).Lock()
	defer (*m.lock).Unlock()

	err := m.db.Put([]byte(key), []byte(value), nil)
	if err != nil {
		return err
	}
	m.data[key] = value

	return nil
}

// Delete performs non-blocking delete of an object from the vault.
func (m LDBMap) Delete(key string) error {
	(*m.lock).Lock()
	defer (*m.lock).Unlock()

	err := m.db.Delete([]byte(key), nil)
	if err != nil {
		return err
	}
	delete(m.data, key)

	return nil
}

// Exist performs non-blocking check of an object for existence in the vault.
func (m LDBMap) Exist(key string) bool {
	(*m.lock).RLock()
	defer (*m.lock).RUnlock()

	_, ok := m.data[key]

	return ok
}

// BatchPut puts an object into the vault by batching (leveldb).
func (m LDBMap) BatchPut(key string, value string) {
	(*m.lock).Lock()
	defer (*m.lock).Unlock()

	m.batch.Put([]byte(key), []byte(value))
}

// BatchDelete deletes an object from the vault by batching (leveldb).
func (m LDBMap) BatchDelete(key string) {
	(*m.lock).Lock()
	defer (*m.lock).Unlock()

	m.batch.Delete([]byte(key))
}

// BatchWrite performs write of the batch to the database.
func (m LDBMap) BatchWrite() error {
	err := m.db.Write(m.batch, nil)

	return err
}

// Iterate brings possibility to iterate over the vault.
func (m LDBMap) Iterate() map[string]string {
	return m.data
}

// Len returns length of the vault
func (m LDBMap) Len() int {
	return len(m.data)
}

// Close closes the database.
func (m LDBMap) Close() {
	m.db.Close()
}
