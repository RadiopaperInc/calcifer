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
	"strings"
	"sync"
)

var (
	defaultFieldCache = newFieldCache()
)

type fieldCache struct {
	cache sync.Map // from reflect.Type to cacheValue
}

func newFieldCache() *fieldCache {
	return &fieldCache{}
}

func MustRegisterModel(m ReadableModel) {
	if err := RegisterModel(m); err != nil {
		panic(err)
	}
}

func RegisterModel(m ReadableModel) error {
	_, err := defaultFieldCache.fields(reflect.TypeOf(m))
	return err
}

type field struct {
	Name        string       // effective field name
	NameFromTag bool         // did Name come from a tag?
	Type        reflect.Type // field type
	Index       []int        // index sequence, for reflect.Value.FieldByIndex
	TagOptions  *tagOptions  // additional options set on the tag

	nameBytes []byte
	equalFold func(s, t []byte) bool
}
type fieldList []field

type cacheValue struct {
	fields fieldList
	err    error
}

func validate(t reflect.Type) error {
	return nil
	// return errors.New("validate: unimplemented")
}

func (c *fieldCache) fields(t reflect.Type) (fieldList, error) {
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("calcifer: fields of non-struct type %q", t.String())
	}
	var cv cacheValue
	x, ok := c.cache.Load(t)
	if ok {
		cv = x.(cacheValue)
	} else {
		if err := validate(t); err != nil {
			cv = cacheValue{nil, err}
		} else {
			f, err := c.typeFields(t)
			cv = cacheValue{fieldList(f), err}
		}
		c.cache.Store(t, cv)
	}
	return cv.fields, cv.err
}

func (c *fieldCache) typeFields(t reflect.Type) ([]field, error) {
	fields, err := listFields(t)
	if err != nil {
		return nil, err
	}

	fs := make(map[string]struct{})
	for _, f := range fields {
		if _, ok := fs[f.Name]; ok {
			return nil, fmt.Errorf("calcifer: duplicate field %q", f.Name)
		}
		fs[f.Name] = struct{}{}
	}
	return fields, nil
}

func listFields(t reflect.Type) ([]field, error) {
	// This uses the same condition that the Go language does: there must be a unique instance
	// of the match at a given depth level. If there are multiple instances of a match at the
	// same depth, they annihilate each other and inhibit any possible match at a lower level.
	// The algorithm is breadth first search, one depth level at a time.

	// The current and next slices are work queues:
	// current lists the fields to visit on this depth level,
	// and next lists the fields on the next lower level.
	current := []fieldScan{}
	next := []fieldScan{{typ: t}}

	// nextCount records the number of times an embedded type has been
	// encountered and considered for queueing in the 'next' slice.
	// We only queue the first one, but we increment the count on each.
	// If a struct type T can be reached more than once at a given depth level,
	// then it annihilates itself and need not be considered at all when we
	// process that next depth level.
	var nextCount map[reflect.Type]int

	// visited records the structs that have been considered already.
	// Embedded pointer fields can create cycles in the graph of
	// reachable embedded types; visited avoids following those cycles.
	// It also avoids duplicated effort: if we didn't find the field in an
	// embedded type T at level 2, we won't find it in one at level 4 either.
	visited := map[reflect.Type]bool{}

	var fields []field // Fields found.

	for len(next) > 0 {
		current, next = next, current[:0]
		count := nextCount
		nextCount = nil

		// Process all the fields at this depth, now listed in 'current'.
		// The loop queues embedded fields found in 'next', for processing during the next
		// iteration. The multiplicity of the 'current' field counts is recorded
		// in 'count'; the multiplicity of the 'next' field counts is recorded in 'nextCount'.
		for _, scan := range current {
			t := scan.typ
			if visited[t] {
				// We've looked through this type before, at a higher level.
				// That higher level would shadow the lower level we're now at,
				// so this one can't be useful to us. Ignore it.
				continue
			}
			visited[t] = true
			for i := 0; i < t.NumField(); i++ {
				f := t.Field(i)

				exported := (f.PkgPath == "")

				// If a named field is unexported, ignore it. An anonymous
				// unexported field is processed, because it may contain
				// exported fields, which are visible.
				if !exported && !f.Anonymous {
					continue
				}

				// Examine the tag.
				tagName, keep, options, err := parseTag(f.Tag)
				if err != nil {
					return nil, err
				}
				if !keep {
					continue
				}
				if isLeafType(f.Type) {
					fields = append(fields, newField(f, tagName, options, scan.index, i))
					continue
				}

				var ntyp reflect.Type
				if f.Anonymous {
					// Anonymous field of type T or *T.
					ntyp = f.Type
					if ntyp.Kind() == reflect.Ptr {
						ntyp = ntyp.Elem()
					}
				}

				// Record fields with a tag name, non-anonymous fields, or
				// anonymous non-struct fields.
				if tagName != "" || ntyp == nil || ntyp.Kind() != reflect.Struct {
					if !exported {
						continue
					}
					fields = append(fields, newField(f, tagName, options, scan.index, i))
					if count[t] > 1 {
						// If there were multiple instances, add a second,
						// so that the annihilation code will see a duplicate.
						fields = append(fields, fields[len(fields)-1])
					}
					continue
				}

				// Queue embedded struct fields for processing with next level,
				// but only if the embedded types haven't already been queued.
				if nextCount[ntyp] > 0 {
					nextCount[ntyp] = 2 // exact multiple doesn't matter
					continue
				}
				if nextCount == nil {
					nextCount = map[reflect.Type]int{}
				}
				nextCount[ntyp] = 1
				if count[t] > 1 {
					nextCount[ntyp] = 2 // exact multiple doesn't matter
				}
				var index []int
				index = append(index, scan.index...)
				index = append(index, i)
				next = append(next, fieldScan{ntyp, index})
			}
		}
	}
	return fields, nil
}

func newField(f reflect.StructField, tagName string, options *tagOptions, index []int, i int) field {
	name := tagName
	if name == "" {
		name = f.Name
	}
	sf := field{
		Name:        name,
		NameFromTag: tagName != "",
		Type:        f.Type,
		TagOptions:  options,
		nameBytes:   []byte(name),
	}
	sf.equalFold = foldFunc(sf.nameBytes)
	sf.Index = append(sf.Index, index...)
	sf.Index = append(sf.Index, i)
	return sf
}

// A fieldScan represents an item on the fieldByNameFunc scan work list.
type fieldScan struct {
	typ   reflect.Type
	index []int
}

func isLeafType(t reflect.Type) bool {
	return t == typeOfGoTime /*|| t == typeOfLatLng || t == typeOfProtoTimestamp */
}

type tagOptions struct {
	omitEmpty       bool   // do not marshal value if empty
	serverTimestamp bool   // set time.Time to server timestamp on write
	reference       string // collection referenced by this field
}

// parseTag interprets firestore struct field tags.
func parseTag(t reflect.StructTag) (name string, keep bool, options *tagOptions, err error) {
	name, keep, opts, err := parseStandardTag("calcifer", t)
	if err != nil {
		return "", false, nil, fmt.Errorf("calcifer: %v", err)
	}
	tagOpts := tagOptions{}
	for _, opt := range opts {
		if strings.HasPrefix(opt, "ref:") {
			tagOpts.reference = strings.TrimPrefix(opt, "ref:")
			continue
		}
		switch opt {
		case "omitempty":
			tagOpts.omitEmpty = true
		case "serverTimestamp":
			tagOpts.serverTimestamp = true
		default:
			return "", false, nil, fmt.Errorf("firestore: unknown tag option: %q", opt)
		}
	}
	return name, keep, &tagOpts, nil
}

// parseStandardTag extracts the sub-tag named by key, then parses it using the
// de facto standard format introduced in encoding/json:
//   "-" means "ignore this tag". It must occur by itself. (parseStandardTag returns an error
//       in this case, whereas encoding/json accepts the "-" even if it is not alone.)
//   "<name>" provides an alternative name for the field
//   "<name>,opt1,opt2,..." specifies options after the name.
// The options are returned as a []string.
func parseStandardTag(key string, t reflect.StructTag) (name string, keep bool, options []string, err error) {
	s := t.Get(key)
	parts := strings.Split(s, ",")
	if parts[0] == "-" {
		if len(parts) > 1 {
			return "", false, nil, errors.New(`"-" field tag with options`)
		}
		return "", false, nil, nil
	}
	if len(parts) > 1 {
		options = parts[1:]
	}
	return parts[0], true, options, nil
}
