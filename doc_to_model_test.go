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

func TestDataToValueMap(t *testing.T) {
	t.Skip("populateMap: unimplemented")
	var m map[string]string
	d := map[string]interface{}{"a": "A", "b": "B"}
	assert.NoError(t, dataToValue(reflect.ValueOf(&m), d))
	assert.Equal(t, d, m)
}

func TestDataToValueSlice(t *testing.T) {
	var s []string
	d := []string{"a", "b", "c"}
	assert.NoError(t, dataToValue(reflect.ValueOf(&s), d))
	assert.Equal(t, d, s)
}

func TestDataToValueStruct(t *testing.T) {
	type testModel struct {
		Model
		Name string `calcifer:"name"`
		ELO  int    `calcifer:"elo_score"`
	}
	var s testModel
	d := map[string]interface{}{"id": "1", "name": "Dave", "elo_score": 2500}
	assert.NoError(t, dataToValue(reflect.ValueOf(&s), d))
	assert.Equal(t, "1", s.ID)
	assert.Equal(t, "Dave", s.Name)
	assert.Equal(t, 2500, s.ELO)
}

func TestDataToValueRelatedStruct(t *testing.T) {
	type relatedModel struct {
		Model
		X int `calcifer:"x"`
	}
	type testModel struct {
		Model
		Rel      relatedModel            `calcifer:"rel,ref:foo"`
		RelSlice []relatedModel          `calcifer:"relslice,ref:foo"`
		RelMap   map[string]relatedModel `calcifer:"relmap,ref:foo"`
	}
	var s testModel
	d := map[string]interface{}{
		"id": "1", "rel": "2",
		"relslice": []any{"3", "4"},
		"relmap":   map[string]any{"five": "5", "six": "6"},
	}
	assert.NoError(t, dataToValue(reflect.ValueOf(&s), d))
	assert.Equal(t, "1", s.ID)
	assert.Equal(t, "2", s.Rel.ID)
	assert.Equal(t, []relatedModel{{Model{ID: "3"}, 0}, {Model{ID: "4"}, 0}}, s.RelSlice)
	assert.Equal(t, map[string]relatedModel{"five": {Model{ID: "5"}, 0}, "six": {Model{ID: "6"}, 0}}, s.RelMap)
}
