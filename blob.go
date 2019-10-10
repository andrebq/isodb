package isodb

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
	"sort"

	"github.com/pkg/errors"
)

type (
	// Blob holds raw bytes without any know structure
	Blob struct {
		// Content of this blob
		Content []byte
	}

	// BlobRefList implements a sorted slice of BlobRef objects. Sorted by Alg and Value
	BlobRefList []BlobRef

	// BlobRef holds the reference to any blob
	BlobRef struct {
		// Algorithm used to compute the hash
		Alg HashAlg

		// Value of the hash itself
		Value string
	}

	// HashAlg lists the available hash algorithms to use
	HashAlg string
)

const (
	// Sha256 indicates Sha256 algorithm
	Sha256 = HashAlg("sha256")
)

func (ha HashAlg) valid() bool {
	switch ha {
	case Sha256:
		return true
	}
	return false
}

func (ha HashAlg) newHash() hash.Hash {
	// TODO(andre): actually use a pool of hash
	return sha256.New()
}

func (ha HashAlg) dispose(h hash.Hash) {
	// TODO(andre): actually dispose to the pool
}

// ComputeBytes the hash using the given algorithm. It is just a syntatic sugar to ComputeReader
func (ha HashAlg) ComputeBytes(r []byte) (BlobRef, error) {
	return ha.ComputeReader(bytes.NewBuffer(r))
}

// ComputeReader the hash using the given algorithm and returns a BlobRef with the right values
func (ha HashAlg) ComputeReader(r io.Reader) (BlobRef, error) {
	if !ha.valid() {
		return BlobRef{}, ErrInvalidHashAlgorithm
	}
	h := ha.newHash()
	defer ha.dispose(h)
	h.Reset()
	_, err := io.Copy(h, r)
	if err != nil {
		return BlobRef{}, errors.Wrapf(err, "isodb: unable to compute hash %v for reader", ha)
	}
	buf := make([]byte, 0, h.Size())
	return BlobRef{
		Alg:   ha,
		Value: base64.RawURLEncoding.EncodeToString(h.Sum(buf)),
	}, nil
}

// NewBlobString returns a new blob from the given string
func NewBlobString(str string) Blob {
	return Blob{
		Content: []byte(str),
	}
}

// ToBlob returns itself
func (b Blob) ToBlob() Blob {
	return b
}

// Ref returns the default hash version of this blob, it is just a syntatic sugar for RefAlg
func (b Blob) Ref() BlobRef {
	ref, err := b.RefAlg(Sha256)
	if err != nil {
		panic("if this is happening there is something really really wrong and we should abort! " + err.Error())
	}
	return ref
}

// RefAlg returns the hash of this blob using the given RefAlg
func (b Blob) RefAlg(h HashAlg) (BlobRef, error) {
	return h.ComputeBytes(b.Content)
}

// ToBlob returns the encoded version of this BlobRef
func (b BlobRef) ToBlob() Blob {
	blob, err := defaultCodec.encode(b)
	if err != nil {
		panic("This should never ever happen! " + err.Error())
	}
	return blob
}

// FromBlob decodes the BlobRef from this Blob
func (b *BlobRef) FromBlob(in Blob) error {
	err := defaultCodec.decode(b, in)
	if err != nil {
		panic("This should never ever happen! " + err.Error())
	}
	return err
}

// IsZero returns true if BlobRef is empty
func (b BlobRef) IsZero() bool {
	return b == (BlobRef{})
}

func (b BlobRef) String() string {
	return fmt.Sprintf("%v:%v", b.Alg, b.Value)
}

func (b BlobRef) less(o BlobRef) bool {
	return b.Alg < o.Alg && b.Value < o.Value
}

// Contains returns true if the parent is present
func (bl BlobRefList) Contains(r BlobRef) bool {
	for _, v := range bl {
		if v == r {
			return true
		}
	}
	return false
}

// SortInPlace this blob ref list
func (bl BlobRefList) SortInPlace() {
	if bl == nil {
		return
	}
	sort.Sort(bl)
}

// Insert the ref in the list or do nothing if the item is there already.
//
// Returns true if a new item was added
func (bl *BlobRefList) Insert(r BlobRef) bool {
	for _, v := range *bl {
		if v == r {
			return false
		}
	}
	*bl = append(*bl, r)
	return true
}

func (bl BlobRefList) Len() int           { return len(bl) }
func (bl BlobRefList) Less(i, j int) bool { return bl[i].less(bl[j]) }
func (bl BlobRefList) Swap(i, j int)      { bl[i], bl[j] = bl[j], bl[i] }
