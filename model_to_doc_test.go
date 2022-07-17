// Copyright 2022 Radiopaper Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
