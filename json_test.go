package x_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/tozd/go/errors"

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

func TestSaveJSONToDir(t *testing.T) {
	t.Parallel()

	type doc struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		errE := x.SaveJSONToDir(context.Background(), dir, []doc{}, func(d doc) (string, errors.E) {
			return d.ID, nil
		})
		assert.NoError(t, errE, "% -+#.1v", errE) //nolint:testifylint

		entries, err := os.ReadDir(dir)
		require.NoError(t, err)
		assert.Empty(t, entries)
	})

	t.Run("saves multiple documents", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		data := []doc{
			{ID: "a", Name: "Alice"},
			{ID: "b", Name: "Bob"},
		}
		errE := x.SaveJSONToDir(context.Background(), dir, data, func(d doc) (string, errors.E) {
			return d.ID, nil
		})
		assert.NoError(t, errE, "% -+#.1v", errE) //nolint:testifylint

		contentA, err := os.ReadFile(filepath.Join(dir, "a.json")) //nolint:gosec
		require.NoError(t, err)
		assert.JSONEq(t, `{"id":"a","name":"Alice"}`, string(contentA))
		assert.True(t, strings.HasSuffix(string(contentA), "\n"))

		contentB, err := os.ReadFile(filepath.Join(dir, "b.json")) //nolint:gosec
		require.NoError(t, err)
		assert.JSONEq(t, `{"id":"b","name":"Bob"}`, string(contentB))
	})

	t.Run("creates directory if missing", func(t *testing.T) {
		t.Parallel()

		dir := filepath.Join(t.TempDir(), "sub", "dir")
		data := []doc{{ID: "c", Name: "Charlie"}}
		errE := x.SaveJSONToDir(context.Background(), dir, data, func(d doc) (string, errors.E) {
			return d.ID, nil
		})
		assert.NoError(t, errE, "% -+#.1v", errE) //nolint:testifylint

		content, err := os.ReadFile(filepath.Join(dir, "c.json")) //nolint:gosec
		require.NoError(t, err)
		assert.JSONEq(t, `{"id":"c","name":"Charlie"}`, string(content))
	})

	t.Run("cancelled context", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		dir := t.TempDir()
		data := []doc{{ID: "d", Name: "Dave"}}
		errE := x.SaveJSONToDir(ctx, dir, data, func(d doc) (string, errors.E) {
			return d.ID, nil
		})
		assert.Error(t, errE)
		assert.ErrorIs(t, errE, context.Canceled)
	})

	t.Run("indented output", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		data := []doc{{ID: "e", Name: "Eve"}}
		errE := x.SaveJSONToDir(context.Background(), dir, data, func(d doc) (string, errors.E) {
			return d.ID, nil
		})
		assert.NoError(t, errE, "% -+#.1v", errE) //nolint:testifylint

		content, err := os.ReadFile(filepath.Join(dir, "e.json")) //nolint:gosec
		require.NoError(t, err)
		expected := "{\n  \"id\": \"e\",\n  \"name\": \"Eve\"\n}\n"
		assert.Equal(t, expected, string(content))
	})

	t.Run("filename error", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		data := []doc{{ID: "f", Name: "Frank"}}
		errE := x.SaveJSONToDir(context.Background(), dir, data, func(_ doc) (string, errors.E) {
			return "", errors.New("bad filename")
		})
		assert.Error(t, errE)
		assert.EqualError(t, errE, "bad filename")
	})

	t.Run("marshal error", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		data := []chan int{make(chan int)}
		errE := x.SaveJSONToDir(context.Background(), dir, data, func(_ chan int) (string, errors.E) {
			return "bad", nil
		})
		assert.Error(t, errE)
	})
}
