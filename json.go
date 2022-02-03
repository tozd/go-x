package x

import (
	"bytes"
	"encoding/json"

	"gitlab.com/tozd/go/errors"
)

// UnmarshalWithoutUnknownFields is a standard JSON unmarshal, just
// that it returns an error if there is any unknown field present in JSON.
func UnmarshalWithoutUnknownFields(data []byte, v interface{}) errors.E {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(v)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// MarshalWithoutEscapeHTML is a standard JSON marshal, just that
// it does not escape HTML characters.
func MarshalWithoutEscapeHTML(v interface{}) ([]byte, errors.E) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}
