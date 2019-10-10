package isodb

type (
	strErr string
)

const (
	// ErrInvalidHashAlgorithm indicates a invalid value for HashAlg
	ErrInvalidHashAlgorithm = strErr("isodb: invalid hash algorithm")

	errNothingChanged = strErr("isodb:internal: nothing changed")
)

func (s strErr) Error() string {
	return string(s)
}
