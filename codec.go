package isodb

import "encoding/json"

type (
	codec struct {
		marshal   func(interface{}) ([]byte, error)
		unmarshal func([]byte, interface{}) error
	}
)

var (
	jsonCodec = codec{
		marshal:   json.Marshal,
		unmarshal: json.Unmarshal,
	}

	defaultCodec = jsonCodec
)

func (c *codec) encode(in interface{}) (Blob, error) {
	buf, err := c.marshal(in)
	return Blob{Content: buf}, err
}

func (c *codec) decode(out interface{}, in Blob) error {
	return c.unmarshal(in.Content, out)
}
