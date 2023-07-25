package x

import (
	"bytes"
	"encoding/json"
	"io"

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
	_, err = decoder.Token()
	if err == nil || !errors.Is(err, io.EOF) {
		return errors.New("invalid data after top-level value")
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
	b := buf.Bytes()
	if len(b) > 0 {
		return b[:len(b)-1], nil
	}
	return b, nil
}
