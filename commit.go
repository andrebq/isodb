package isodb

type (
	// Commit represents a single snapshot of the entire database
	Commit struct {
		// Folder points to the root direction of this commit
		Folder BlobRef

		// Parents points to the list of previous commit before this one
		Parents BlobRefList
	}
)

// ToBlob encodes this Commit object as a Blob object for future use/reference
func (c *Commit) ToBlob() Blob {
	blob, err := defaultCodec.encode(c)
	if err != nil {
		panic("this should never ever happen! " + err.Error())
	}
	return blob
}

// FromBlob updates Commit from the Blob object
func (c *Commit) FromBlob(b Blob) error {
	var tmp Commit
	err := defaultCodec.decode(&tmp, b)
	if err != nil {
		return err
	}
	*c = tmp
	return nil
}
