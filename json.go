package x

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"gitlab.com/tozd/go/errors"
)

var ErrJSONUnmarshalExtraData = errors.Base("invalid data after top-level value")

// UnmarshalWithoutUnknownFields is a standard JSON unmarshal, just
// that it returns an error if there is any unknown field present in JSON.
// It also adds JSON data to any error as an error detail.
func UnmarshalWithoutUnknownFields(data []byte, v interface{}) errors.E {
	errE := DecodeJSONWithoutUnknownFields(bytes.NewReader(data), v)
	if errE != nil {
		errors.Details(errE)["json"] = string(data)
	}
	return errE
}

// MarshalWithoutEscapeHTML is a standard JSON marshal, just that
// it does not escape HTML characters.
func MarshalWithoutEscapeHTML(v interface{}) ([]byte, errors.E) {
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)
	if err != nil {
		return nil, errors.WithStack(err)
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
		return errors.WithDetails(err, "json", string(data))
	}
	return nil
}

// Marshal is a standard JSON marshal, just
// that it returns an error with a stack trace.
func Marshal(v interface{}) ([]byte, errors.E) {
	b, err := json.Marshal(v)
	return b, errors.WithStack(err)
}

// DecodeJSON reads one JSON object from reader and unmarshals it into v.
// It errors if there is more data trailing after the object.
func DecodeJSON(reader io.Reader, v interface{}) errors.E {
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(v)
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = decoder.Token()
	if !errors.Is(err, io.EOF) {
		return errors.WithStack(ErrJSONUnmarshalExtraData)
	}
	return nil
}

// DecodeJSONWithoutUnknownFields reads one JSON object from reader and
// unmarshals it into v.
// It errors if there is more data trailing after the object.
// It returns an error if there is any unknown field present in JSON.
func DecodeJSONWithoutUnknownFields(reader io.Reader, v interface{}) errors.E {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(v)
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = decoder.Token()
	if !errors.Is(err, io.EOF) {
		return errors.WithStack(ErrJSONUnmarshalExtraData)
	}
	return nil
}

// Same as gitlab.com/tozd/go/zerolog.TimeFieldFormat.
const timeFieldFormat = "2006-01-02T15:04:05.000Z07:00"

// Time is the same as [time.Time], only that it marshals to JSON with millisecond
// precision to minimize any side channels.
type Time time.Time

// MarshalJSON implements [json.Marshaler] interface for Time.
func (t Time) MarshalJSON() ([]byte, error) {
	return MarshalWithoutEscapeHTML(time.Time(t).Format(timeFieldFormat))
}

// UnmarshalJSON implements [json.Marshaler] interface for Time.
func (t *Time) UnmarshalJSON(data []byte) error {
	var tt time.Time
	errE := UnmarshalWithoutUnknownFields(data, &tt)
	if errE != nil {
		return errE
	}
	*t = Time(tt)
	return nil
}

// Duration is the same as [time.Duration], only that it marshals to JSON as string.
type Duration time.Duration

// MarshalJSON implements [json.Marshaler] interface for Duration.
func (d Duration) MarshalJSON() ([]byte, error) {
	return MarshalWithoutEscapeHTML(time.Duration(d).String())
}

// UnmarshalJSON implements [json.Marshaler] interface for Duration.
func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	errE := UnmarshalWithoutUnknownFields(b, &s)
	if errE != nil {
		return errE
	}
	tmp, err := time.ParseDuration(s)
	if err != nil {
		return errors.WithDetails(err, "duration", s)
	}
	*d = Duration(tmp)
	return nil
}
