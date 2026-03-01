package x_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

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

func TestUnmarshal(t *testing.T) {
	t.Parallel()

	type Test struct {
		Field string `json:"field"`
	}

	var v Test
	errE := x.Unmarshal([]byte(`{"field":"abc"}`), &v)
	require.NoError(t, errE, "% -+#.1v", errE)
	assert.Equal(t, "abc", v.Field)

	errE = x.Unmarshal([]byte(`invalid json`), &v)
	assert.Error(t, errE)
}

func TestMarshal(t *testing.T) {
	t.Parallel()

	type Test struct {
		Field string `json:"field"`
	}

	data, errE := x.Marshal(Test{Field: "abc"})
	require.NoError(t, errE, "% -+#.1v", errE)
	assert.Equal(t, `{"field":"abc"}`, string(data)) //nolint:testifylint
}

func TestDecodeJSON(t *testing.T) {
	t.Parallel()

	type Test struct {
		Field string `json:"field"`
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		var v Test
		errE := x.DecodeJSON(strings.NewReader(`{"field":"abc"}`), &v)
		require.NoError(t, errE, "% -+#.1v", errE)
		assert.Equal(t, "abc", v.Field)
	})

	t.Run("extra data", func(t *testing.T) {
		t.Parallel()

		var v Test
		errE := x.DecodeJSON(strings.NewReader(`{"field":"abc"} extra`), &v)
		assert.Error(t, errE)
		assert.ErrorIs(t, errE, x.ErrJSONUnmarshalExtraData)
	})

	t.Run("decode error", func(t *testing.T) {
		t.Parallel()

		var v Test
		errE := x.DecodeJSON(strings.NewReader(`invalid`), &v)
		assert.Error(t, errE)
	})
}

func TestTime(t *testing.T) {
	t.Parallel()

	now := time.Now().Truncate(time.Millisecond).UTC()
	xt := x.Time(now)

	data, err := json.Marshal(xt)
	require.NoError(t, err)

	var xt2 x.Time
	err = json.Unmarshal(data, &xt2)
	require.NoError(t, err)
	assert.Equal(t, now, time.Time(xt2).UTC())

	t.Run("unmarshal error", func(t *testing.T) {
		t.Parallel()

		var xt3 x.Time
		err := json.Unmarshal([]byte(`"not-a-time"`), &xt3)
		assert.Error(t, err)
	})
}

func TestDuration(t *testing.T) {
	t.Parallel()

	d := x.Duration(5 * time.Second)

	data, err := json.Marshal(d)
	require.NoError(t, err)
	assert.Equal(t, `"5s"`, string(data))

	var d2 x.Duration
	err = json.Unmarshal(data, &d2)
	require.NoError(t, err)
	assert.Equal(t, d, d2)

	t.Run("invalid duration", func(t *testing.T) {
		t.Parallel()

		var d3 x.Duration
		err := json.Unmarshal([]byte(`"invalid"`), &d3)
		assert.Error(t, err)
	})

	t.Run("non-string duration", func(t *testing.T) {
		t.Parallel()

		var d4 x.Duration
		err := json.Unmarshal([]byte(`123`), &d4)
		assert.Error(t, err)
	})
}
