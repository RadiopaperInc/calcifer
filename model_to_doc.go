package calcifer

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

func modelToDoc(m ReadableModel) (interface{}, error) {
	v := reflect.ValueOf(m)
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
	return nil, nil // TODO: errors.New("calcifer: sliceToInterface: unimplemented")
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
		if f.TagOptions.reference != "" {
			fk, err := valueToForeignKey(v.FieldByIndex(f.Index))
			if err != nil {
				return nil, err
			}
			sm[f.Name] = fk
		} else {
			val, err := valueToInterface(v.FieldByIndex(f.Index))
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
	return sv.String(), nil
}
