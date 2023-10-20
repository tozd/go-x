package x

import (
	"fmt"
	"reflect"

	"gitlab.com/tozd/go/errors"
)

var ErrNotFoundInStruct = errors.Base("not found in struct")

// FindInStruct returns a pointer to the field of type T found in struct value.
// value can be of type T by itself and in that case it is simply returned.
func FindInStruct[T any](value interface{}) (*T, errors.E) {
	// TODO: Replace with reflect.TypeFor.
	typeToGet := reflect.TypeOf((*T)(nil)).Elem()
	val := reflect.ValueOf(value).Elem()
	typ := val.Type()
	if typ == typeToGet {
		return val.Addr().Interface().(*T), nil //nolint:forcetypeassert
	}
	fields := reflect.VisibleFields(typ)
	for _, field := range fields {
		if field.Type == typeToGet {
			return val.FieldByIndex(field.Index).Addr().Interface().(*T), nil //nolint:forcetypeassert
		}
	}

	errE := errors.WithDetails( //nolint:forcetypeassert
		ErrNotFoundInStruct,
		"getType", fmt.Sprintf("%T", *reflect.ValueOf(new(T)).Interface().(*T)),
		"valueType", fmt.Sprintf("%T", value),
	)
	return nil, errE
}
