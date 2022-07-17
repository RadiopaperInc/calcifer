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
	"context"
	"errors"
	"reflect"
)

func (c *Client) expandModel(ctx context.Context, m MutableModel) error {
	v := reflect.ValueOf(m)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil // no model to expand
		}
		v = v.Elem()
	}
	fs, err := defaultFieldCache.fields(v.Type())
	if err != nil {
		return err
	}
	for _, f := range fs {
		if f.TagOptions.reference == "" {
			continue
		}
		rv := v.FieldByIndex(f.Index)
		if rv.Kind() != reflect.Pointer {
			return errors.New("calcifier: trying to expand into non-pointer field")
		}
		sv := rv.Elem().FieldByName("Model") // TODO: ensure this is a calcifer.Model?
		if sv.Kind() != reflect.Struct {
			return errors.New("calcifer: missing Model field on foreign key reference object")
		}
		sv = sv.FieldByName("ID")
		id := sv.String()
		if id == "" {
			continue // empty field, no ID to expand
		}
		ref := c.Collection(f.TagOptions.reference).Doc(id)
		if err := ref.Get(ctx, rv.Interface().(MutableModel)); err != nil {
			return err
		}
	}
	return nil
}
