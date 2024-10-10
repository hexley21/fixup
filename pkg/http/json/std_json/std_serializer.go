package std_json

import (
	"encoding/json"
	"errors"
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
	var ute *json.UnmarshalTypeError
	if errors.As(err, &ute) {
		return fmt.Errorf("unmarshal type error: expected=%v, got=%v, field=%v, offset=%v", ute.Type, ute.Value, ute.Field, ute.Offset)
	}
	var se *json.SyntaxError
	if errors.As(err, &se) {
		return fmt.Errorf("syntax error: offset=%v, error=%v", se.Offset, se.Error())
	}
	return err
}
