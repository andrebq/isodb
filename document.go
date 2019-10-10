package isodb

import "github.com/segmentio/ksuid"

type (
	// DocumentKey contains the identification of a document
	DocumentKey struct {
		// Set to group documents into a collection
		Set string
		// Id of the document (32-bit k-sorted ids)
		K ksuid.KSUID
	}
)

// NewRandomKey returns a new random key for the given set
func NewRandomKey(set string) DocumentKey {
	return DocumentKey{
		Set: set,
		K:   ksuid.New(),
	}
}

func (d DocumentKey) paths() []string {
	var p []string
	str := d.K.String()
	p = append(p, d.Set, str[:2], str[2:4], str[4:6], str[6:8], str[8:10], str[10:12], str)
	return p
}
