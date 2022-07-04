package calcifer

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDataToValueBool(t *testing.T) {
	var b bool
	assert.NoError(t, dataToValue(reflect.ValueOf(&b), true))
	assert.True(t, b)
	assert.NoError(t, dataToValue(reflect.ValueOf(&b), false))
	assert.False(t, b)
}

func TestDataToValueString(t *testing.T) {
	var s string
	assert.NoError(t, dataToValue(reflect.ValueOf(&s), "Hello, world!"))
	assert.Equal(t, "Hello, world!", s)
}

func TestDataToValueInt(t *testing.T) {
	var i int
	assert.NoError(t, dataToValue(reflect.ValueOf(&i), int(42)))
	assert.Equal(t, 42, i)
}

func TestDataToValueTime(t *testing.T) {
	var ts time.Time
	now := time.Now()
	want := now
	assert.NoError(t, dataToValue(reflect.ValueOf(&ts), now))
	assert.True(t, want.Equal(now))
}

func TestDataToValuePointer(t *testing.T) {
	v := 42
	var i int
	assert.NoError(t, dataToValue(reflect.ValueOf(&i), &v))
	assert.Equal(t, 42, i)
}

func TestDataPointerToValue(t *testing.T) {
	var i *int
	assert.NoError(t, dataToValue(reflect.ValueOf(&i), int(42)))
	assert.Equal(t, 42, *i)
}
