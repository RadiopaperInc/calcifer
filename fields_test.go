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

	_, err = defaultFieldCache.fields(reflect.TypeOf(&e))
	assert.NoError(t, err)
}
