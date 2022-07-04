package calcifer

import (
	"fmt"
	"reflect"
)

var (
	defaultFieldCache = newFieldCache()
)

type fieldCache struct {
}

func newFieldCache() *fieldCache {
	return &fieldCache{}
}

func MustRegisterModel(m ReadableModel) {
	if err := RegisterModel(m); err != nil {
		panic(err)
	}
}

func RegisterModel(m ReadableModel) error {
	_, err := defaultFieldCache.fields(reflect.TypeOf(m))
	return err
}

type field struct{}
type fieldList []field

func (c *fieldCache) fields(t reflect.Type) (fieldList, error) {
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("calcifer: fields of non-struct type %q", t.String())
	}
	// TODO: extract fields from struct, and parse tags
	return nil, nil
}
