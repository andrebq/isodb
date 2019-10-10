package isodb

import (
	"bytes"
	"io"
)

type (
	// KV abstraction used to perform atomic operations. Empty keys are not allowed.
	KV interface {
		io.Closer

		// PutNew if the key isn't found in the database
		PutNew(k string, b Blob) (bool, error)

		// Put the given key in the database
		Put(k string, b Blob) error

		// PutIf updates k if the old version passes the given check function
		PutIf(k string, b Blob, check CheckFn) (bool, error)

		// CAS update the key if old value matches the expected old (syntaic sugar for PutIf)
		CAS(k string, old, new Blob) (bool, error)

		// Get returns the value for the given key
		Get(k string) (Blob, error)

		// Has returns if the key is present in the database
		Has(k string) (bool, error)
	}

	// CheckFn is by PutIf
	CheckFn func(prev, next Blob) (bool, error)
)

const (
	// ErrCASNotExecuted indicates that a KV CAS operation didn't work
	ErrCASNotExecuted = strErr("isodb: unable to perform CAS operation")
)

func alwaysTrue(_, _ Blob) (bool, error) { return true, nil }
func onlyIfMissing(old, _ Blob) (bool, error) {
	return len(old.Content) == 0, nil
}
func cas(old Blob) CheckFn {
	return func(prev, next Blob) (bool, error) {
		return bytes.Equal(prev.Content, next.Content), nil
	}
}
