package x_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/tozd/go/errors"

	"gitlab.com/tozd/go/x"
)

type BaseStruct struct {
	Name string
}

type ExtendedStruct struct {
	BaseStruct
	Age int
}

func TestFindInStruct(t *testing.T) {
	t.Parallel()

	y, err := x.FindInStruct[BaseStruct](&BaseStruct{Name: "x"})
	assert.NoError(t, err)
	assert.Equal(t, "x", y.Name)

	y, err = x.FindInStruct[BaseStruct](&ExtendedStruct{BaseStruct: BaseStruct{Name: "x"}, Age: 30})
	assert.NoError(t, err)
	assert.Equal(t, "x", y.Name)

	z, err := x.FindInStruct[ExtendedStruct](&ExtendedStruct{BaseStruct: BaseStruct{Name: "x"}, Age: 30})
	assert.NoError(t, err)
	assert.Equal(t, "x", z.Name)
	assert.Equal(t, 30, z.Age)

	_, err = x.FindInStruct[ExtendedStruct](&BaseStruct{Name: "x"})
	assert.ErrorIs(t, err, x.ErrNotFoundInStruct)
	assert.Equal(t, "x_test.ExtendedStruct", errors.AllDetails(err)["getType"])
	assert.Equal(t, "*x_test.BaseStruct", errors.AllDetails(err)["valueType"])

	name, err := x.FindInStruct[string](&BaseStruct{Name: "x"})
	assert.NoError(t, err)
	assert.Equal(t, "x", *name)

	s := ExtendedStruct{BaseStruct: BaseStruct{Name: "x"}, Age: 30}
	age, err := x.FindInStruct[int](&s)
	assert.NoError(t, err)
	assert.Equal(t, 30, *age)

	*age = 31
	assert.Equal(t, 31, s.Age)
}
