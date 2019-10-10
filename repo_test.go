package isodb

import (
	"bytes"
	"testing"
)

func newRepo(t *testing.T) *Repo {
	kv, err := NewTempKV()
	if err != nil {
		t.Fatal(err)
	}
	repo := NewRepoWithKV(kv)
	return repo
}

func TestRepo(t *testing.T) {
	repo := newRepo(t)
	cs := NewChangeset()
	bob := NewRandomKey("people")
	alice := NewRandomKey("alice")
	cs.Put(bob, NewBlobString("bob bobson"))
	cs.Put(alice, NewBlobString("alice anderson"))

	ref, err := repo.Apply(cs)
	if err != nil {
		t.Fatal(err)
	}

	err = repo.UpdatePointer("master/HEAD", ref, BlobRef{})
	if err != nil {
		t.Fatal(err)
	}

	if ptrRef, err := repo.GetPointer("master/HEAD"); err != nil {
		t.Fatal(err)
	} else if ref != ptrRef {
		t.Fatalf("Should have updated pointer. Expecting %v got %v", ref, ptrRef)
	}

	commit, err := repo.GetCommit(ref)
	if err != nil {
		t.Fatal(err)
	}

	cs = NewChangeset(commit.ToBlob().Ref())
	cs.Put(bob, NewBlobString("Bob Buffon"))

	nextRef, err := repo.Apply(cs)
	if err != nil {
		t.Fatal(err)
	}
	nextCommit, err := repo.GetCommit(nextRef)
	if err != nil {
		t.Fatal(err)
	}

	if !nextCommit.Parents.Contains(ref) {
		t.Fatalf("Commit %v should be a children of %v", nextRef.String(), ref.String())
	}

	content, err := repo.GetContentAtKey(ref, bob)
	if err != nil {
		t.Fatal(err)
	} else if !bytes.Equal(content.Content, NewBlobString("bob bobson").Content) {
		t.Fatal("Content not found")
	}

	content, err = repo.GetContentAtKey(nextRef, bob)
	if err != nil {
		t.Fatal(err)
	} else if !bytes.Equal(content.Content, NewBlobString("Bob Buffon").Content) {
		t.Fatalf("Content differs. Got %v", content.Content)
	}

	content, err = repo.GetContentAtKey(nextRef, alice)
	if err != nil {
		t.Fatal(err)
	} else if !bytes.Equal(content.Content, NewBlobString("alice anderson").Content) {
		t.Fatalf("Content differs. Got %v", content.Content)
	}
}
