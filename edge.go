package isodb

import "sort"

type (
	// Edge contains a name and a link to a blob
	Edge struct {
		Name string
		Ref  BlobRef
	}

	// EdgeList is a list of edges sorted by Name
	EdgeList []Edge
)

// SortInPlace this list, always calls this before serializing to keep the hash consistent
func (el EdgeList) SortInPlace() {
	sort.Sort(el)
}

// FindByName the edge with the given name and returns it (if found), its index.
//
// Returns -1 if the edge isn't found
func (el EdgeList) FindByName(name string) (Edge, Idx) {
	for i, v := range el {
		if v.Name == name {
			return v, Idx(i)
		}
	}
	return Edge{}, IdxNotFound
}

// Insert will update this EdgeList to include the given edge, returns the updated list if the edge is new
// or the same list if the edge already exists.
//
// The array is sorted after this operation regardless of it being changed or not.
func (el EdgeList) Insert(e Edge) (EdgeList, Idx) {
	el.SortInPlace()
	for i, v := range el {
		if v.Name == e.Name {
			(el)[i] = e
			return el, Idx(i)
		}
	}
	copied := make(EdgeList, 0, len(el))
	copied = append(copied, el...)
	copied = append(copied, e)
	copied.SortInPlace()

	return copied, IdxNotFound
}

func (el EdgeList) Less(i, j int) bool { return el[i].Name < el[j].Name }
func (el EdgeList) Len() int           { return len(el) }
func (el EdgeList) Swap(i, j int)      { el[i], el[j] = el[j], el[i] }
