package isodb

import (
	"io/ioutil"

	badger "github.com/dgraph-io/badger"
)

type (
	// boltKV
	boltKV struct {
		db *badger.DB
	}
)

// NewTempKV returns a kv-implementation using a temporary folder
func NewTempKV() (KV, error) {
	dir, err := ioutil.TempDir("", "repo")
	if err != nil {
		return nil, err
	}
	return NewPersistentKV(dir)
}

// NewPersistentKV returns a kv-implementation using badger
func NewPersistentKV(folder string) (KV, error) {
	db, err := badger.Open(badger.DefaultOptions(folder).WithLogger(nil))
	if err != nil {
		return nil, err
	}
	return &boltKV{db: db}, nil
}

// Put implements KV
func (bdb *boltKV) Put(k string, b Blob) error {
	_, e := bdb.PutIf(k, b, alwaysTrue)
	return e
}

// PutIf implements KV
func (bdb *boltKV) PutIf(k string, b Blob, fn CheckFn) (bool, error) {
	var change bool
	bk := []byte(k)
	err := bdb.db.Update(func(tx *badger.Txn) error {
		item, err := tx.Get(bk)
		if err == badger.ErrKeyNotFound {
			item = nil
		} else if err != nil {
			return err
		}

		if item == nil {
			change, err = fn(Blob{}, b)
		} else {
			err = item.Value(func(v []byte) error {
				change, err = fn(Blob{Content: v}, b)
				if err != nil {
					return err
				}
				return nil
			})
		}

		if err != nil {
			return err
		}
		if !change {
			return errNothingChanged
		}
		return tx.Set(bk, b.Content)
	})
	if err == errNothingChanged {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return change, nil
}

// CAS implements KV
func (bdb *boltKV) CAS(k string, old, new Blob) (bool, error) {
	return bdb.PutIf(k, new, cas(old))
}

// Get return the value for the given k
func (bdb *boltKV) Get(k string) (Blob, error) {
	var b Blob
	bk := []byte(k)
	err := bdb.db.View(func(tx *badger.Txn) error {
		item, err := tx.Get(bk)
		if err != nil {
			return err
		}
		if item == nil {
			return nil
		}
		b.Content, err = item.ValueCopy(nil)
		return err
	})
	return b, err
}

// Close implements KV
func (bdb *boltKV) Close() error {
	return bdb.db.Close()
}

// Has implements KV
func (bdb *boltKV) Has(k string) (bool, error) {
	var size int64
	bk := []byte(k)
	err := bdb.db.View(func(tx *badger.Txn) error {
		item, err := tx.Get(bk)
		if err == badger.ErrKeyNotFound {
			return nil
		} else if err != nil {
			return err
		} else if item == nil {
			return nil
		}
		size = item.ValueSize()
		return nil
	})
	return size > 0, err
}

// PutNew implements KV
func (bdb *boltKV) PutNew(k string, b Blob) (bool, error) {
	return bdb.PutIf(k, b, onlyIfMissing)
}
