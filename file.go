package isodb

type (
	// File contains links to sub-folders or files
	File struct {
		Name     string
		Leaf     bool
		Children EdgeList
	}
)

// GetFileContent returns the blob edge link for the given file or empty.
func (f *File) GetFileContent() BlobRef {
	edge, idx := f.Children.FindByName("blob")
	if idx.NotFound() {
		return BlobRef{}
	}
	return edge.Ref
}

// SetFileContent creates a copy of this file with a different blob content
func (f *File) SetFileContent(ref BlobRef) *File {
	updated := *f
	updated.Children, _ = f.Children.Insert(Edge{Name: "blob", Ref: ref})
	return &updated
}

// ToBlob encodes this File object as a Blob object for future use/reference
func (f *File) ToBlob() Blob {
	blob, err := defaultCodec.encode(f)
	if err != nil {
		panic("this should never ever happen! " + err.Error())
	}
	return blob
}

// FromBlob updates File from the Blob object
func (f *File) FromBlob(b Blob) error {
	var tmp File
	err := defaultCodec.decode(&tmp, b)
	if err != nil {
		return err
	}
	*f = tmp
	return nil
}

// Add the Edge to the file and return a new entry
func (f *File) Add(children Edge) *File {
	updated := *f
	updated.Children, _ = f.Children.Insert(children)
	return &updated
}
