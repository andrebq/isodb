package isodb

type (
	// Idx is used to add extra methods to integers representing idx of things
	Idx int
)

const (
	// IdxNotFound indicates when a value wasn't found in the array
	IdxNotFound = Idx(-1)
)

// NotFound returns true if the value wasn't found in the slice
func (i Idx) NotFound() bool {
	return i == IdxNotFound
}
