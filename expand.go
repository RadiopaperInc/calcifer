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
	"fmt"
	"reflect"

	"cloud.google.com/go/firestore"
	"golang.org/x/sync/errgroup"
)

func (c *Client) expandField(ctx context.Context, rv reflect.Value, col string) error {
	if rv.Kind() == reflect.Slice {
		for i := 0; i < rv.Len(); i++ { // TODO: parallelize
			el := rv.Index(i)
			if el.Kind() == reflect.Pointer {
				if el.IsNil() {
					el.Set(reflect.New(el.Type().Elem()))
				}
				el = el.Elem()
			}
			if err := c.expandField(ctx, el.Addr(), col); err != nil {
				return err
			}
		}
		return nil
	} else if rv.Kind() == reflect.Map {
		return errors.New("caclifer: expansion of maps to foreign keys unimplemented")
	} else if rv.Kind() != reflect.Pointer {
		return errors.New("calcifier: trying to expand into non-pointer field")
	}
	if rv.IsNil() {
		return nil
	}
	sv := rv.Elem().FieldByName("Model") // TODO: ensure this is a calcifer.Model?
	if sv.Kind() != reflect.Struct {
		return errors.New("calcifer: missing Model field on foreign key reference object")
	}
	sv = sv.FieldByName("ID")
	id := sv.String()
	if id == "" {
		return nil // empty field, no ID to expand
	}
	ref := c.Collection(col).Doc(id)
	if err := ref.Get(ctx, rv.Interface().(MutableModel)); err != nil {
		return err
	}
	return nil
}

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
	for _, f := range fs { // TODO: parallelize
		col := f.TagOptions.reference
		if col == "" {
			continue
		}
		rv := v.FieldByIndex(f.Index)
		if err := c.expandField(ctx, rv, col); err != nil {
			return err
		}
	}
	return nil
}

func (tx *Transaction) expandField(rv reflect.Value, col string) error {
	if rv.Kind() == reflect.Slice {
		for i := 0; i < rv.Len(); i++ { // TODO: parallelize
			el := rv.Index(i)
			if el.Kind() == reflect.Pointer {
				if el.IsNil() {
					el.Set(reflect.New(el.Type().Elem()))
				}
				el = el.Elem()
			}
			if err := tx.expandField(el.Addr(), col); err != nil {
				return err
			}
		}
		return nil
	} else if rv.Kind() == reflect.Map {
		return errors.New("caclifer: expansion of maps to foreign keys unimplemented")
	} else if rv.Kind() != reflect.Pointer {
		return errors.New("calcifier: trying to expand into non-pointer field")
	}
	if rv.IsNil() {
		return nil
	}
	sv := rv.Elem().FieldByName("Model") // TODO: ensure this is a calcifer.Model?
	if sv.Kind() != reflect.Struct {
		return errors.New("calcifer: missing Model field on foreign key reference object")
	}
	sv = sv.FieldByName("ID")
	id := sv.String()
	if id == "" {
		return nil // empty field, no ID to expand
	}
	ref := tx.cli.Collection(col).Doc(id)
	return tx.Get(ref, rv.Interface().(MutableModel))
}

func (tx *Transaction) expandModel(m MutableModel) error {
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
	for _, f := range fs { // TODO: parallelize
		col := f.TagOptions.reference
		if col == "" {
			continue
		}
		rv := v.FieldByIndex(f.Index)
		if err := tx.expandField(rv, col); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) expandAll(ctx context.Context, p any) error {
	modelSlice := reflect.ValueOf(p).Elem()
	fs, err := defaultFieldCache.fields(modelSlice.Index(0).Type())
	if err != nil {
		return err
	}
	refs := make([][]*firestore.DocumentRef, len(fs))
	for fi, f := range fs {
		if f.TagOptions.reference == "" {
			continue
		}
		refs[fi] = make([]*firestore.DocumentRef, modelSlice.Len())
		for i := 0; i < modelSlice.Len(); i++ {
			rv := modelSlice.Index(i).FieldByIndex(fs[fi].Index)
			sv := rv.Elem().FieldByName("Model") // TODO: ensure this is a calcifer.Model?
			if sv.Kind() != reflect.Struct {
				return errors.New("calcifer: missing Model field on foreign key reference object")
			}
			sv = sv.FieldByName("ID")
			id := sv.String()
			if id == "" {
				continue // empty field, no ID to expand | TODO, ok if we GetAll with this in the slice?
			}
			refs[fi][i] = c.Collection(f.TagOptions.reference).cref.Doc(id)
		}
	}
	g, gctx := errgroup.WithContext(ctx)
	for fi, f := range fs {
		if f.TagOptions.reference == "" {
			continue
		}
		fi := fi
		g.Go(func() error {
			docs, err := c.fs.GetAll(gctx, refs[fi])
			if err != nil {
				return err
			}
			for i, doc := range docs {
				if refs[fi][i] != nil {
					if !doc.Exists() {
						return fmt.Errorf("calcifer: unable to find doc with ID %q during expansion of collection %q", refs[fi][i].ID, fs[fi].TagOptions.reference)
					}
					if err := docToModel(modelSlice.Index(i).FieldByIndex(fs[fi].Index).Interface().(MutableModel), doc); err != nil {
						return err
					}
				}
			}
			return nil
		})
	}
	return g.Wait()
}
