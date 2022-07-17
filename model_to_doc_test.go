package calcifer

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValueToInterfaceBool(t *testing.T) {
	b := true
	i, err := valueToInterface(reflect.ValueOf(b))
	assert.NoError(t, err)
	assert.Equal(t, true, i.(bool))

	b = false
	i, err = valueToInterface(reflect.ValueOf(b))
	assert.NoError(t, err)
	assert.Equal(t, false, i.(bool))
}

func TestValueToInterfaceString(t *testing.T) {
	s := "Hello, world!"
	i, err := valueToInterface(reflect.ValueOf(s))
	assert.NoError(t, err)
	assert.Equal(t, "Hello, world!", i.(string))
}

func TestValueToInterfaceInt(t *testing.T) {
	n := 42
	i, err := valueToInterface(reflect.ValueOf(n))
	assert.NoError(t, err)
	assert.Equal(t, int64(42), i.(int64))
}

// TODO: struct, etc.
