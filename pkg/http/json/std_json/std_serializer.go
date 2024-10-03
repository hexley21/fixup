package std_json

import (
	"encoding/json"
	"fmt"
	"io"
)

type stdJSONSerializer struct{}

func New() *stdJSONSerializer {
	return &stdJSONSerializer{}
}

func (j *stdJSONSerializer) Serialize(w io.Writer, i any) error {
	return json.NewEncoder(w).Encode(i)
}

func (j *stdJSONSerializer) Deserialize(reader io.Reader, i any) error {
	err := json.NewDecoder(reader).Decode(i)
	if ute, ok := err.(*json.UnmarshalTypeError); ok {
		return fmt.Errorf("unmarshal type error: expected=%v, got=%v, field=%v, offset=%v", ute.Type, ute.Value, ute.Field, ute.Offset)
	} else if se, ok := err.(*json.SyntaxError); ok {
		return fmt.Errorf("syntax error: offset=%v, error=%v", se.Offset, se.Error())
	}
	return err
}
