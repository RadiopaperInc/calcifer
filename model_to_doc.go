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
)

func modelToDoc(m ReadableModel) (interface{}, error) {
	v := reflect.ValueOf(m)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil, nil
		}
		v = v.Elem()
	}
	_, err := defaultFieldCache.fields(v.Type())
	if err != nil {
		return nil, err
	}
	return structToInterface(v)
}

func valueToInterface(v reflect.Value) (interface{}, error) {
	vi := v.Interface()
	switch x := vi.(type) {
	case []byte:
		return x, nil
	case time.Time:
		return x, nil
	}
	switch v.Kind() {
	case reflect.Bool:
		return v.Bool(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), nil
	case reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return v.Uint(), nil
	case reflect.Float32, reflect.Float64:
		return v.Float(), nil
	case reflect.String:
		return v.String(), nil
	case reflect.Slice:
		return sliceToInterface(v)
	case reflect.Map:
		return mapToInterface(v)
	case reflect.Struct:
		return structToInterface(v)
	case reflect.Ptr:
		if v.IsNil() {
			return nil, nil
		}
		return valueToInterface(v.Elem())
	case reflect.Interface:
		if v.NumMethod() == 0 { // empty interface: recurse on its contents
			return valueToInterface(v.Elem())
		}
		fallthrough // any other interface value is an error

	default:
		return nil, fmt.Errorf("calcifer: cannot convert type %s to firestore value", v.Type())
	}
}

func sliceToInterface(v reflect.Value) (interface{}, error) {
	st := v.Type().Elem()
	switch st.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		st = typeOfInt64
	case reflect.Uint8, reflect.Uint16, reflect.Uint32:
		st = typeOfUInt64
	case reflect.Float32, reflect.Float64:
		st = typeOfFloat64
	}
	st = reflect.SliceOf(st)
	sv := reflect.MakeSlice(st, v.Len(), v.Len())
	for i := 0; i < v.Len(); i++ {
		iv, err := valueToInterface(v.Index(i))
		if err != nil {
			return nil, err
		}
		sv.Index(i).Set(reflect.ValueOf(iv))
	}
	return sv.Interface(), nil
}

func mapToInterface(v reflect.Value) (interface{}, error) {
	return nil, errors.New("calcifer: mapToInterface: unimplemented")
}

func structToInterface(v reflect.Value) (interface{}, error) {
	fs, err := defaultFieldCache.fields(v.Type())
	if err != nil {
		return nil, err
	}
	sm := make(map[string]interface{})
	for _, f := range fs {
		fv := v.FieldByIndex(f.Index)
		if f.TagOptions.reference != "" {
			if fv.Kind() == reflect.Slice {
				fk, err := valueToForeignKeySlice(fv)
				if err != nil {
					return nil, err
				}
				sm[f.Name] = fk
			} else if fv.Kind() == reflect.Map {
				fk, err := valueToForeignKeyMap(fv)
				if err != nil {
					return nil, err
				}
				sm[f.Name] = fk
			} else {
				fk, err := valueToForeignKey(fv)
				if err != nil {
					return nil, err
				}
				sm[f.Name] = fk
			}
		} else {
			val, err := valueToInterface(fv)
			if err != nil {
				return nil, err
			}
			sm[f.Name] = val
		}
	}
	return sm, nil
}

func valueToForeignKey(v reflect.Value) (string, error) {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return "", nil
		}
		return valueToForeignKey(v.Elem())
	}
	if v.Kind() != reflect.Struct {
		return "", errors.New("caclifer: cannot use non-struct type as foreign key reference")
	}
	// TODO validate the struct embeds Model
	sv := v.FieldByName("Model") // TODO: ensure this is a calcifer.Model?
	if sv.Kind() != reflect.Struct {
		return "", errors.New("calcifer: missing Model field on foreign key reference object")
	}
	sv = sv.FieldByName("ID")
	ss := sv.String()
	if ss == "" {
		return "", errors.New("calcifer: cannot convert Model to foreign key with empty ID field")
	}
	return ss, nil
}

func valueToForeignKeySlice(v reflect.Value) ([]string, error) {
	n := v.Len()
	fk := make([]string, n)
	for i := 0; i < n; i++ {
		vi := v.Index(i)
		if v.Kind() == reflect.Pointer { // TODO: is this needed?
			v = v.Elem()
		}
		fki, err := valueToForeignKey(vi)
		if err != nil {
			return nil, err
		}
		fk[i] = fki
	}
	return fk, nil
}

func valueToForeignKeyMap(v reflect.Value) (map[string]string, error) {
	fk := make(map[string]string)
	iter := v.MapRange()
	for iter.Next() {
		k := iter.Key()
		if k.Kind() != reflect.String {
			return nil, errors.New("calcifer: keys in foreign-key maps must be strings")
		}
		fkk, err := valueToForeignKey(iter.Value())
		if err != nil {
			return nil, err
		}
		fk[k.String()] = fkk
	}
	return fk, nil
}
