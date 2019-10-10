package isodb

import (
	"errors"
)

type (
	// Repo contains all the commits/changes written to the database
	Repo struct {
		kv KV
	}

	toBlober interface {
		ToBlob() Blob
	}

	blobMap interface {
		put(b toBlober)
		read(out interface{}, r BlobRef) bool
		raw(out []byte, r BlobRef) Blob
		has(BlobRef) bool
		keys() []BlobRef
	}
)

const (
	// ErrInvalidOldRef the expected value for the pointer is old
	ErrInvalidOldRef = strErr("isodb: invalid old reference")

	// ErrDocumentNotFound document not found
	ErrDocumentNotFound = strErr("isodb: document not found")
)

// NewPersistentRepo returns a new Repo with a persistent stored in `folder`.
func NewPersistentRepo(folder string) (*Repo, error) {
	kv, err := NewPersistentKV(folder)
	if err != nil {
		return nil, err
	}
	return &Repo{kv: kv}, nil
}

// NewRepoWithKV returns a new Repo using the given KV
func NewRepoWithKV(kv KV) *Repo {
	return &Repo{kv: kv}
}

// UpdatePointer ptr from oldRef to newRef, if oldRef is empty then it will only update
// if value is new.
//
// If the change cannot be executed, then ErrInvalidOldRef is returned.
func (r *Repo) UpdatePointer(ptr string, newRef, oldRef BlobRef) error {
	ptr = "refs/" + ptr
	if oldRef.IsZero() {
		ok, err := r.kv.PutNew(ptr, newRef.ToBlob())
		if err != nil {
			return err
		} else if !ok {
			return ErrInvalidOldRef
		}
	}
	ok, err := r.kv.CAS(ptr, oldRef.ToBlob(), newRef.ToBlob())
	if err != nil {
		return err
	} else if !ok {
		return ErrInvalidOldRef
	}
	return nil
}

// GetPointer returns the ref from the given pointer
func (r *Repo) GetPointer(ptr string) (BlobRef, error) {
	ptr = "refs/" + ptr
	val, err := r.kv.Get(ptr)
	if err != nil {
		return BlobRef{}, err
	}
	var br BlobRef
	return br, br.FromBlob(val)
}

// GetBlob returns the blob from the given BlobRef
func (r *Repo) GetBlob(ref BlobRef) (Blob, error) {
	return r.kv.Get(ref.String())
}

// GetCommit returns the Commit pointed by BlobRef
func (r *Repo) GetCommit(ref BlobRef) (Commit, error) {
	var c Commit
	b, err := r.GetBlob(ref)
	if err != nil {
		return Commit{}, err
	}
	return c, c.FromBlob(b)
}

// GetFile returns the file pointed by BlobRef
func (r *Repo) GetFile(ref BlobRef) (File, error) {
	var f File
	b, err := r.GetBlob(ref)
	if err != nil {
		return File{}, err
	}
	return f, f.FromBlob(b)
}

// GetContentAtKey returns the blob at the given key or null if they key does not exist
func (r *Repo) GetContentAtKey(commitRef BlobRef, key DocumentKey) (Blob, error) {
	c, err := r.GetCommit(commitRef)
	if err != nil {
		return Blob{}, err
	}
	file, err := r.GetFile(c.Folder)
	if err != nil {
		return Blob{}, err
	}

	steps := key.paths()
	for _, p := range steps {
		e, i := file.Children.FindByName(p)
		if i.NotFound() {
			return Blob{}, ErrDocumentNotFound
		}
		file, err = r.GetFile(e.Ref)
		if err != nil {
			return Blob{}, err
		}
	}

	blobEdge, i := file.Children.FindByName("blob")
	if i.NotFound() {
		return Blob{}, ErrDocumentNotFound
	}

	return r.GetBlob(blobEdge.Ref)
}

// Apply the provided Changeset to the repository and returns the reference to the new commit
func (r *Repo) Apply(cs *Changeset) (BlobRef, error) {
	cs.parents.SortInPlace()
	cs.ensureLeafs()
	var root *File
	switch len(cs.parents) {
	case 0:
		root = &File{}
	case 1:
		parent, err := r.GetCommit(cs.parents[0])
		if err != nil {
			return BlobRef{}, err
		}
		rootFile, err := r.GetFile(parent.Folder)
		if err != nil {
			return BlobRef{}, err
		}
		root = &rootFile
	default:
		return BlobRef{}, errors.New("isodb: cannot handle merge commits yet! sorry 😅")
	}
	if len(cs.parents) > 1 {
		return BlobRef{}, errors.New("cannot handle merge commits at this point. sorry about that")
	}
	blobs := &kvBlobMap{cache: &inMemBlobMap{}, kv: r.kv}

	for k, v := range cs.leafs {
		steps := k.paths()

		blobs.put(v)
		thisRoot := addPathToLeaf(steps, v.Ref(), blobs)
		root = mergeRoots(root, thisRoot, blobs)
		blobs.put(root)
	}
	c := Commit{
		Folder:  root.ToBlob().Ref(),
		Parents: cs.parents,
	}
	blobs.put(&c)
	return c.ToBlob().Ref(), r.persistCommit(c, blobs)
}

// actually store the blobs in the underlying database starting from the commit to the children node
func (r *Repo) persistCommit(c Commit, b blobMap) error {
	for _, k := range b.keys() {
		blob := b.raw(nil, k)
		_, err := r.kv.PutNew(k.String(), blob)
		if err != nil {
			return err
		}
	}
	return nil
}

func addPathToLeaf(steps []string, blobRef BlobRef, blobs blobMap) *File {
	leafFile := &File{
		Name: steps[len(steps)-1],
		Leaf: true,
	}
	leafFile = leafFile.SetFileContent(blobRef)

	steps = steps[:len(steps)-1]
	blobs.put(leafFile)

	for i := len(steps) - 1; i >= 0; i-- {
		f := &File{
			Name:     steps[i],
			Children: EdgeList{Edge{Name: leafFile.Name, Ref: leafFile.ToBlob().Ref()}},
		}
		blobs.put(f)
		leafFile = f
	}

	root := &File{
		Name:     "root",
		Children: EdgeList{Edge{Name: leafFile.Name, Ref: leafFile.ToBlob().Ref()}},
	}
	blobs.put(root)
	return root
}

// merge entries from partialRoot into full root and returns fullRoot.
//
// nothing is updated in place
//
// TODO(andre): think of a way to avoid soo much allocation here
func mergeRoots(fullRoot, partialRoot *File, blobs blobMap) *File {
	if len(partialRoot.Children) == 0 {
		// there is nothing to merge, just ignore it then
		return fullRoot
	}
	switch {
	case fullRoot.Leaf && partialRoot.Leaf:
		// we reached the leaf File on both roots
		// copy the partialRoot (new content) to fullRoot (old content)
		fullRoot = fullRoot.SetFileContent(partialRoot.GetFileContent())
		return fullRoot
	}
	nextChildrenEdge := partialRoot.Children[0]
	oldChildrenEdge, idx := fullRoot.Children.FindByName(nextChildrenEdge.Name)
	if idx.NotFound() {
		fullRoot = fullRoot.Add(nextChildrenEdge)
		return fullRoot
	}

	// old Children exists, now we have to merge them before returning
	var nextChildrenFile File
	if !blobs.read(&nextChildrenFile, nextChildrenEdge.Ref) {
		panic("blobs does not have " + nextChildrenEdge.Ref.String())
	}

	var oldChildrenFile File
	if !blobs.read(&oldChildrenFile, oldChildrenEdge.Ref) {
		panic("blobs does not have " + oldChildrenEdge.Ref.String())
	}

	f := mergeRoots(&oldChildrenFile, &nextChildrenFile, blobs)
	blobs.put(f)
	return fullRoot.Add(Edge{Name: oldChildrenEdge.Name, Ref: f.ToBlob().Ref()})
}
