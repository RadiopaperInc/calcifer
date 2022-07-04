package calcifer

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterModel(t *testing.T) {
	assert.NoError(t, RegisterModel(Event{}))
}

func TestFields(t *testing.T) {
	m := make(map[string]interface{})
	_, err := defaultFieldCache.fields(reflect.TypeOf(m))
	assert.Error(t, err)

	e := Event{}
	_, err = defaultFieldCache.fields(reflect.TypeOf(e))
	assert.NoError(t, err)
}
