package rps

import (
	"path/filepath"
	"strconv"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
)

// UMap structure.
type UMap struct {
	data  map[int64]*User
	lock  *sync.RWMutex
	name  string
	db    *leveldb.DB
	batch *leveldb.Batch
}

// NewUMap creates an object of UMap structure.
func NewUMap(name string, dbPath string) UMap {
	lock := sync.RWMutex{}
	data := map[int64]*User{}
	db, err := leveldb.OpenFile(filepath.Join(dbPath, name), nil)
	if err != nil {
		Error.Printf("Can't get access to database:\n\tDBPath: %s",
			filepath.Join(dbPath, name))
		panic(err)
	}

	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		_k, _v := iter.Key(), iter.Value()
		uid, err := strconv.ParseInt(string(_k), 10, 64)
		if err != nil {
			Error.Printf("Can't parse user ID:\n\tUserID: %s",
				string(_k))
			panic(err)
		}
		user, err := Deserialize(_v)
		if err != nil {
			Error.Printf("Can't deserialize user:\n\tUserID: %d",
				uid)
			panic(err)
		}
		data[uid] = &user
	}

	batch := new(leveldb.Batch)
	UMap := UMap{data, &lock, name, db, batch}

	return UMap
}

// Get performs non-blocking get of an object from the vault.
func (m UMap) Get(uid int64) *User {
	(*m.lock).RLock()
	defer (*m.lock).RUnlock()
	return m.data[uid]
}

// Put performs non-blocking put of an object into the vault.
func (m UMap) Put(uid int64, user *User) error {
	(*m.lock).Lock()
	defer (*m.lock).Unlock()

	strUID := strconv.FormatInt(uid, 10)
	strUser := user.Serialize()

	err := m.db.Put([]byte(strUID), strUser, nil)
	if err != nil {
		return err
	}
	m.data[uid] = user

	return nil
}

// Delete performs non-blocking delete of an object from the vault.
func (m UMap) Delete(uid int64) error {
	(*m.lock).Lock()
	defer (*m.lock).Unlock()

	strUID := strconv.FormatInt(uid, 10)

	err := m.db.Delete([]byte(strUID), nil)
	if err != nil {
		return err
	}
	delete(m.data, uid)

	return nil
}

// Exist performs non-blocking check of an object for existence in the vault.
func (m UMap) Exist(uid int64) bool {
	(*m.lock).RLock()
	defer (*m.lock).RUnlock()

	_, ok := m.data[uid]

	return ok
}

// BatchPut puts an object into the vault by batching (leveldb).
func (m UMap) BatchPut(uid int64, user *User) {
	(*m.lock).Lock()
	defer (*m.lock).Unlock()

	strUID := strconv.FormatInt(uid, 10)
	strUser := user.Serialize()

	m.batch.Put([]byte(strUID), strUser)
}

// BatchDelete deletes an object from the vault by batching (leveldb).
func (m UMap) BatchDelete(uid int64) {
	(*m.lock).Lock()
	defer (*m.lock).Unlock()

	strUID := strconv.FormatInt(uid, 10)

	m.batch.Delete([]byte(strUID))
}

// BatchWrite performs write of the batch to the database.
func (m UMap) BatchWrite() error {
	err := m.db.Write(m.batch, nil)

	return err
}

// Iterate brings possibility to iterate over the vault.
func (m UMap) Iterate() map[int64]*User {
	return m.data
}

// Len returns length of the vault
func (m UMap) Len() int {
	return len(m.data)
}

// Close closes the database.
func (m UMap) Close() {
	m.db.Close()
}
