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
		return errors.New("calcifer: slice deserialization unimplemented")
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
				if err := dataToValue(v.FieldByIndex(f.Index), dd); err != nil {
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
