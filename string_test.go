package x_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/tozd/go/x"
)

func TestString2ByteSlice(t *testing.T) {
	t.Parallel()

	s := "hello"
	b := x.String2ByteSlice(s)
	assert.Equal(t, []byte("hello"), b)
	assert.Len(t, b, 5)
}

func TestByteSlice2String(t *testing.T) {
	t.Parallel()

	b := []byte("hello")
	s := x.ByteSlice2String(b)
	assert.Equal(t, "hello", s)

	// Empty slice should return empty string.
	assert.Equal(t, "", x.ByteSlice2String(nil))
	assert.Equal(t, "", x.ByteSlice2String([]byte{}))
}
