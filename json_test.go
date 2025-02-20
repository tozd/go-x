package x_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/tozd/go/x"
)

func TestUnmarshalWithoutUnknownFields(t *testing.T) {
	t.Parallel()

	type Test struct {
		Field string
	}

	var v Test

	err := x.UnmarshalWithoutUnknownFields([]byte(`{}`), &v)
	assert.NoError(t, err, "% -+#.1v", err) //nolint:testifylint

	err = x.UnmarshalWithoutUnknownFields([]byte(`{"field": "abc"}`), &v)
	assert.NoError(t, err, "% -+#.1v", err) //nolint:testifylint

	err = x.UnmarshalWithoutUnknownFields([]byte(`{"field2": "abc"}`), &v)
	assert.Error(t, err)

	// Extra payload should error.
	// See: https://github.com/golang/go/issues/36225
	err = x.UnmarshalWithoutUnknownFields([]byte(`{"field": "abc"} xxx`), &v)
	assert.Error(t, err)
	err = x.UnmarshalWithoutUnknownFields([]byte(`{"field": "abc"} ]`), &v)
	assert.Error(t, err)
	err = x.UnmarshalWithoutUnknownFields([]byte(`{"field": "abc"} {`), &v)
	assert.Error(t, err)

	// Extra whitespace should not error.
	err = x.UnmarshalWithoutUnknownFields([]byte(`{"field": "abc"} `), &v)
	assert.NoError(t, err, "% -+#.1v", err)
}

func TestMarshalWithoutEscapeHTML(t *testing.T) {
	t.Parallel()

	type Test struct {
		Field string `json:"field"`
	}

	data, err := x.MarshalWithoutEscapeHTML(&Test{Field: "<body></body>"})
	require.NoError(t, err, "% -+#.1v", err)
	assert.Equal(t, `{"field":"<body></body>"}`, string(data)) //nolint:testifylint
}
