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

func TestValueToInterfaceTime(t *testing.T) {
	now := time.Now()
	ts, err := valueToInterface(reflect.ValueOf(now))
	assert.NoError(t, err)
	assert.Equal(t, now, ts.(time.Time))
}

func TestValueToInterfacePointer(t *testing.T) {
	n := 42
	i, err := valueToInterface(reflect.ValueOf(&n))
	assert.NoError(t, err)
	assert.Equal(t, int64(42), i.(int64))
}

func TestValueToInterfaceMap(t *testing.T) {
	t.Skip("mapToInterface: unimplemented")
	d := map[string]string{"a": "A", "b": "B"}
	i, err := valueToInterface(reflect.ValueOf(d))
	assert.NoError(t, err)
	assert.Equal(t, d, i.(map[string]string))
}

func TestValueToInterfaceStruct(t *testing.T) {
	type coord struct {
		X int `calcifer:"x"`
		Y int `calcifer:"y"`
	}
	c := coord{-3, 7}
	i, err := valueToInterface(reflect.ValueOf(c))
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"x": int64(-3), "y": int64(7)}, i)
}

func TestValueToInterfaceSliceField(t *testing.T) {
	type sliceholder struct {
		X []int `calcifer:"x"`
	}
	s := sliceholder{X: []int{-3, 7}}
	i, err := valueToInterface(reflect.ValueOf(s))
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"x": []int64{-3, 7}}, i)
}

func TestModelToDoc(t *testing.T) {
	type testModel struct {
		Model
		Name string `calcifer:"name"`
		ELO  int    `calcifer:"elo_score"`
	}
	m := testModel{
		Model: Model{ID: "1"},
		Name:  "Dave",
		ELO:   2500,
	}

	i, err := modelToDoc(m)
	assert.NoError(t, err)
	im := i.(map[string]interface{})
	assert.Equal(t, "1", im["id"])
	assert.Equal(t, "Dave", im["name"])
	assert.Equal(t, int64(2500), im["elo_score"])
}

func TestRelatedModelToDoc(t *testing.T) {
	type relatedModel struct {
		Model
		X int `calcifer:"x"`
	}
	type testModel struct {
		Model
		Rel      relatedModel            `calcifer:"rel,ref:foo"`
		RelPtr   *relatedModel           `calcifer:"relptr,ref:foo"`
		RelSlice []relatedModel          `calcifer:"relslice,ref:foo"`
		RelMap   map[string]relatedModel `calcifer:"relmap,ref:foo"`
	}

	m1 := testModel{
		Model:    Model{ID: "1"},
		Rel:      relatedModel{Model: Model{ID: "2"}},
		RelPtr:   &relatedModel{Model: Model{ID: "3"}},
		RelSlice: []relatedModel{{Model: Model{ID: "4"}}, {Model: Model{ID: "5"}}},
		RelMap:   map[string]relatedModel{"six": {Model: Model{ID: "6"}}, "seven": {Model: Model{ID: "7"}}},
	}
	i1, err := modelToDoc(m1)
	assert.NoError(t, err)
	im := i1.(map[string]any)
	assert.Equal(t, "1", im["id"])
	assert.Equal(t, "2", im["rel"])
	assert.Equal(t, "3", im["relptr"])
	assert.Equal(t, []string{"4", "5"}, im["relslice"])
	assert.Equal(t, map[string]string{"six": "6", "seven": "7"}, im["relmap"])

	m2 := testModel{
		Model: Model{ID: "1"},
	}
	i2, err := modelToDoc(m2)
	assert.NoError(t, err)
	im = i2.(map[string]any)
	assert.Equal(t, "1", im["id"])
	assert.Empty(t, im["relptr"])
}
