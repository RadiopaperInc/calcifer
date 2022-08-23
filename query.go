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
	"reflect"

	"cloud.google.com/go/firestore"
)

// Query represents a Firestore query.
//
// Query values are immutable. Each Query method creates
// a new Query; it does not modify the old.
type Query struct {
	cli *Client
	q   firestore.Query
}

// A Queryer is a Query or a CollectionRef. CollectionRefs act as queries whose
// results are all the documents in the collection.
type Queryer interface {
	query() *Query
}

func (q Query) query() *Query {
	return &q
}

// Where returns a new Query that filters the set of results.
// A Query can have multiple filters.
// The path argument can be a single field or a dot-separated sequence of
// fields, and must not contain any of the runes "˜*/[]".
// The op argument must be one of "==", "!=", "<", "<=", ">", ">=",
// "array-contains", "array-contains-any", "in" or "not-in".
func (q Query) Where(path, op string, value interface{}) Query {
	return Query{cli: q.cli, q: q.q.Where(path, op, value)}
}

// OrderBy returns a new Query that specifies the order in which results are
// returned. A Query can have multiple OrderBy/OrderByPath specifications.
// OrderBy appends the specification to the list of existing ones.
//
// The path argument can be a single field or a dot-separated sequence of
// fields, and must not contain any of the runes "˜*/[]".
//
// To order by document name, use the special field path DocumentID.
func (q Query) OrderBy(path string, dir firestore.Direction) Query {
	return Query{cli: q.cli, q: q.q.OrderBy(path, dir)}
}

// Limit returns a new Query that specifies the maximum number of first results
// to return. It must not be negative.
func (q Query) Limit(n int) Query {
	return Query{cli: q.cli, q: q.q.Limit(n)}
}

// LimitToLast returns a new Query that specifies the maximum number of last
// results to return. It must not be negative.
func (q Query) LimitToLast(n int) Query {
	return Query{cli: q.cli, q: q.q.LimitToLast(n)}
}

// TODO: StartAt, StartAfter, EndAt, EndBefore; which require un-wrapping DocumentRefs.

// TODO: Serialize, Deserialize

// Documents returns an iterator over the query's resulting documents.
func (q Query) Documents(ctx context.Context) *DocumentIterator {
	return &DocumentIterator{
		cli: q.cli,
		it:  q.q.Documents(ctx),
	}
}

type DocumentIterator struct {
	cli *Client
	tx  *Transaction
	it  *firestore.DocumentIterator
}

// Next fetches the next result from Firestore, and unmarshals it into p.
// If error is iterator.Done, no result is unmarshalled. Once Next returns Done,
// all subsequent calls will return
// Done.
func (it *DocumentIterator) Next(ctx context.Context, p MutableModel) error {
	doc, err := it.it.Next()
	if err != nil {
		return err
	}
	if err := docToModel(p, doc); err != nil {
		return err
	}

	// TODO: make expansion optional
	expandFunc := it.cli.expandModel
	if it.tx != nil { // expand in the same transaction
		expandFunc = it.tx.cli.expandModel
	}
	if err := expandFunc(ctx, p); err != nil {
		return err
	}

	return nil
}

func (it *DocumentIterator) GetAll(ctx context.Context, p any) error {
	docs, err := it.it.GetAll()
	if err != nil {
		return err
	}

	t := reflect.TypeOf(p).Elem().Elem()
	newSlice := reflect.MakeSlice(reflect.SliceOf(t), len(docs), len(docs))
	reflect.ValueOf(p).Elem().Set(newSlice)

	for i, doc := range docs {
		mm := newSlice.Index(i).Addr().Interface().(MutableModel)
		err := docToModel(mm, doc)
		if err != nil {
			return err
		}
	}
	if len(docs) > 0 {
		expandFunc := it.cli.expandAll
		if it.tx != nil { // expand in the same transaction
			expandFunc = it.tx.cli.expandAll
		}
		if err := expandFunc(ctx, p); err != nil {
			return err
		}
	}
	return nil
}
