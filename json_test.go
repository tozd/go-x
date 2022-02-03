package x_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/tozd/go/x"
)

func TestUnmarshalWithoutUnknownFields(t *testing.T) {
	type Test struct {
		Field string
	}

	var v Test

	err := x.UnmarshalWithoutUnknownFields([]byte(`{}`), &v)
	assert.NoError(t, err)

	err = x.UnmarshalWithoutUnknownFields([]byte(`{"field": "abc"}`), &v)
	assert.NoError(t, err)

	err = x.UnmarshalWithoutUnknownFields([]byte(`{"field2": "abc"}`), &v)
	assert.Error(t, err)
}

func TestMarshalWithoutEscapeHTML(t *testing.T) {
	type Test struct {
		Field string `json:"field"`
	}

	data, err := x.MarshalWithoutEscapeHTML(&Test{Field: "<body></body>"})
	assert.NoError(t, err)
	assert.Equal(t, `{"field":"<body></body>"}`, string(data))
}
