package calcifer

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

var (
	defaultFieldCache = newFieldCache()
)

type fieldCache struct {
	cache sync.Map // from reflect.Type to cacheValue
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

type cacheValue struct {
	fields fieldList
	err    error
}

func validate(t reflect.Type) error {
	return errors.New("validate: unimplemented")
}

func (c *fieldCache) fields(t reflect.Type) (fieldList, error) {
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("calcifer: fields of non-struct type %q", t.String())
	}
	var cv cacheValue
	x, ok := c.cache.Load(t)
	if ok {
		cv = x.(cacheValue)
	} else {
		if err := validate(t); err != nil {
			cv = cacheValue{nil, err}
		} else {
			f, err := c.typeFields(t)
			cv = cacheValue{fieldList(f), err}
		}
		c.cache.Store(t, cv)
	}
	return cv.fields, cv.err
}

func (c *fieldCache) typeFields(t reflect.Type) ([]field, error) {
	return nil, errors.New("typeFields: unimplemented")
}
