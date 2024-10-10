package json

import (
	"io"
)

type Processor interface {
	Serializer
	Deserializer
}

type Serializer interface {
	Serialize(w io.Writer, i any) error
}

type Deserializer interface {
	Deserialize(reader io.Reader, i any) error
}
