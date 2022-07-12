package calcifer

import "reflect"

func modelToDoc(m ReadableModel) (map[string]interface{}, error) {
	v := reflect.ValueOf(m)
	fs, err := defaultFieldCache.fields(v.Type())
	if err != nil {
		return nil, err
	}
	sm := make(map[string]interface{})
	for _, f := range fs {
		// TODO: if field is a struct or map; do this recursively
		sm[f.Name] = v.FieldByIndex(f.Index).Interface()
	}
	return sm, nil
}
