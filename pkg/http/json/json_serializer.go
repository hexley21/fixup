package json

import (
	"io"
)

type JSONProcessor interface {
	JSONSerializer
	JSONDeserializer
}

type JSONSerializer interface {
	Serialize(w io.Writer, i any) error
}

type JSONDeserializer interface {
	Deserialize(reader io.Reader, i any) error
}
