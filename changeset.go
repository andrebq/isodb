package isodb

type (
	// Changeset is used to prepare a commit before actually commiting to it.
	//
	// Changeset shouldn't be shared between different goroutines as it is not a goroutine-safe structre
	Changeset struct {
		// leafs for this changeset, aka, the actual information
		leafs map[DocumentKey]Blob

		// ref of the parent commit
		parents BlobRefList
	}
)

// NewChangeset returns the changeset with the given parents as the previous commit.
//
// It is safe to not provide any parents
func NewChangeset(parents ...BlobRef) *Changeset {
	cs := &Changeset{
		parents: BlobRefList(parents),
	}
	cs.parents.SortInPlace()
	return cs
}

// Put the document in the changeset to be later added to the commit
func (c *Changeset) Put(k DocumentKey, b Blob) {
	c.ensureLeafs()
	c.leafs[k] = b
}

// Read the document in the changeset (only if the document is indexed for changing).
//
// The content is copied to buf and returned as a buf
func (c *Changeset) Read(out []byte, k DocumentKey) (Blob, bool) {
	c.ensureLeafs()
	b, ok := c.leafs[k]
	if !ok {
		return Blob{}, false
	}
	out = append(out, b.Content...)
	return Blob{Content: out}, true
}

func (c *Changeset) ensureLeafs() {
	if c.leafs != nil {
		return
	}
	c.leafs = make(map[DocumentKey]Blob)
}
