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
	"errors"
	"fmt"
	"reflect"
	"time"

	"cloud.google.com/go/firestore"
)

var (
	typeOfByteSlice = reflect.TypeOf([]byte{})
	typeOfGoTime    = reflect.TypeOf(time.Time{})
)

func docToModel(m MutableModel, doc *firestore.DocumentSnapshot) error {
	d := doc.Data()
	v := reflect.ValueOf(m)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return errors.New("calcifer: nil or not a pointer")
	}
	if err := dataToValue(v, d); err != nil {
		return err
	}
	m.setID(doc.Ref.ID)
	m.setCreateTime(doc.CreateTime)
	m.setUpdateTime(doc.UpdateTime)
	return nil
}

func dataToValue(v reflect.Value, d interface{}) error {
	typeErr := func() error {
		return fmt.Errorf("calcifer: cannot set type %s to %s", v.Type(), reflect.TypeOf(d))
	}

	// set nillable types to nil
	if d == nil {
		switch v.Kind() {
		case reflect.Interface, reflect.Ptr, reflect.Map, reflect.Slice:
			v.Set(reflect.Zero(v.Type()))
		}
		return nil
	}

	// dereference data pointers
	dv := reflect.ValueOf(d)
	if dv.Kind() == reflect.Ptr {
		return dataToValue(v, dv.Elem().Interface())
	}

	// convert special types
	switch v.Type() {
	case typeOfGoTime:
		x, ok := d.(time.Time)
		if !ok {
			return typeErr()
		}
		v.Set(reflect.ValueOf(x))
		return nil
	case typeOfByteSlice:
		x, ok := d.([]byte)
		if !ok {
			return typeErr()
		}
		v.SetBytes(x)
		return nil
	}

	// convert supported kinds
	switch v.Kind() {
	case reflect.Ptr:
		// If the pointer is nil, allocate a zero value.
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return dataToValue(v.Elem(), d)
	case reflect.Struct:
		x, ok := d.(map[string]interface{})
		if !ok {
			return typeErr()
		}
		return populateStruct(v, x)
	case reflect.Map:
		x, ok := d.(map[string]interface{})
		if !ok {
			return typeErr()
		}
		return populateMap(v, x)
	case reflect.Slice:
		if dv.Kind() != reflect.Slice {
			return typeErr()
		}
		dlen := dv.Len()
		vlen := v.Len()
		if vlen < dlen {
			v.Set(reflect.MakeSlice(v.Type(), dlen, dlen))
		} else {
			v.SetLen(dlen)
		}
		for i := 0; i < dlen; i++ {
			if err := dataToValue(v.Index(i), dv.Index(i).Interface()); err != nil {
				return err
			}
		}
	case reflect.Bool:
		x, ok := d.(bool)
		if !ok {
			return typeErr()
		}
		v.SetBool(x)
	case reflect.String:
		x, ok := d.(string)
		if !ok {
			return typeErr()
		}
		v.SetString(x)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var i int64
		switch x := d.(type) {
		case int:
			i = int64(x)
		case int8:
			i = int64(x)
		case int16:
			i = int64(x)
		case int32:
			i = int64(x)
		case int64:
			i = int64(x)
		}
		if v.OverflowInt(i) {
			return fmt.Errorf("calcifer: value %v overflows type %s", i, v.Type())
		}
		v.SetInt(i)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var i uint64
		switch x := d.(type) {
		case uint:
			i = uint64(x)
		case uint8:
			i = uint64(x)
		case uint16:
			i = uint64(x)
		case uint32:
			i = uint64(x)
		case uint64:
			i = uint64(x)
		}
		if v.OverflowUint(i) {
			return fmt.Errorf("calcifer: value %v overflows type %s", i, v.Type())
		}
		v.SetUint(i)
	case reflect.Float32, reflect.Float64:
		return errors.New("calcifer: float deserialization unimplemented")
	default:
		return fmt.Errorf("calcifer: cannot set type %s", v.Type())
	}
	return nil
}

func populateStruct(v reflect.Value, d map[string]interface{}) error {
	fs, err := defaultFieldCache.fields(v.Type())
	if err != nil {
		return err
	}
OUTER:
	for k, dd := range d {
		for _, f := range fs { // TODO: make fs hold a map
			if f.Name == k {
				rf := v.FieldByIndex(f.Index)
				if f.TagOptions.reference != "" && dd != nil {
					if ds, ok := dd.([]interface{}); ok {
						if err := populateForeignKeySlice(rf, ds); err != nil {
							return err
						}
					} else if dm, ok := dd.(map[string]interface{}); ok {
						if err := populateForeignKeyMap(rf, dm); err != nil {
							return err
						}
					} else {
						if rf.Kind() == reflect.Pointer {
							if rf.IsNil() {
								rf.Set(reflect.New(rf.Type().Elem()))
							}
							rf = rf.Elem()
						}
						if err := populateForeignKey(rf, dd); err != nil {
							return err
						}
					}
				} else if err := dataToValue(rf, dd); err != nil {
					return err
				}
				continue OUTER
			}
		}
		return fmt.Errorf("calcifer: no struct field matched document field %q", k)
	}
	return nil
}

func populateMap(v reflect.Value, d map[string]interface{}) error {
	return errors.New("calcifer: populateMap: unimplemented")
}

func populateForeignKey(v reflect.Value, dd interface{}) error {
	d, ok := dd.(string)
	if !ok {
		return fmt.Errorf("calcifier: cannot use non-string value %q as foreign key", reflect.TypeOf(dd))
	}
	sv := v.FieldByName("ID")
	if sv.Kind() != reflect.String {
		return errors.New("calcifer: missing string ID field on foreign key model")
	}
	sv.SetString(d)
	return nil
}

func populateForeignKeySlice(v reflect.Value, d []interface{}) error {
	vlen := v.Len()
	dlen := len(d)
	// Make a slice of the right size, avoiding allocation if possible.
	if vlen < dlen {
		v.Set(reflect.MakeSlice(v.Type(), dlen, dlen))
	} else {
		v.SetLen(dlen)
	}

	for i := 0; i < dlen; i++ {
		if err := populateForeignKey(v.Index(i), d[i]); err != nil {
			return err
		}
	}
	return nil
}

func populateForeignKeyMap(v reflect.Value, d map[string]interface{}) error {
	vt := v.Type()
	if v.IsNil() {
		v.Set(reflect.MakeMap(vt))
	}
	et := vt.Elem()
	for k := range d {
		el := reflect.New(et).Elem()
		if err := populateForeignKey(el, d[k]); err != nil {
			return err
		}
		v.SetMapIndex(reflect.ValueOf(k), el)
	}
	return nil
}
