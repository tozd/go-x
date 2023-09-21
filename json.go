package x

import (
	"bytes"
	"encoding/json"
	"io"

	"gitlab.com/tozd/go/errors"
)

var ErrJSONUnmarshalExtraData = errors.Base("invalid data after top-level value")

// UnmarshalWithoutUnknownFields is a standard JSON unmarshal, just
// that it returns an error if there is any unknown field present in JSON.
// It also adds JSON data to any error as an error detail.
func UnmarshalWithoutUnknownFields(data []byte, v interface{}) errors.E {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(v)
	if err != nil {
		errE := errors.WithMessage(err, "json unmarshal")
		errors.Details(errE)["json"] = string(data)
		return errE
	}
	_, err = decoder.Token()
	if err == nil || !errors.Is(err, io.EOF) {
		errE := errors.WithMessage(ErrJSONUnmarshalExtraData, "json unmarshal")
		errors.Details(errE)["json"] = string(data)
		return errE
	}
	return nil
}

// MarshalWithoutEscapeHTML is a standard JSON marshal, just that
// it does not escape HTML characters.
func MarshalWithoutEscapeHTML(v interface{}) ([]byte, errors.E) {
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)
	if err != nil {
		return nil, errors.WithMessage(err, "json marshal")
	}
	b := buf.Bytes()
	if len(b) > 0 {
		// Remove trailing \n which is added by Encode.
		return b[:len(b)-1], nil
	}
	return b, nil
}

// Unmarshal is a standard JSON unmarshal, just
// that it returns an error with a stack trace and
// adds JSON data as an error detail.
func Unmarshal(data []byte, v interface{}) errors.E {
	err := json.Unmarshal(data, v)
	if err != nil {
		errE := errors.WithMessage(err, "json unmarshal")
		errors.Details(errE)["json"] = string(data)
		return errE
	}
	return nil
}

// MarshalWithoutEscapeHTML is a standard JSON marshal, just
// that it returns an error with a stack trace.
func Marshal(v interface{}) ([]byte, errors.E) {
	b, err := json.Marshal(v)
	return b, errors.WithMessage(err, "json marshal")
}
