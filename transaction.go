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

	"cloud.google.com/go/firestore"
)

type Transaction struct {
	tx  *firestore.Transaction
	cli *Client
}

type TransactionOption any

func (c *Client) RunTransaction(ctx context.Context, f func(context.Context, *Transaction) error, opts ...TransactionOption) (err error) {
	return c.fs.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		t := &Transaction{tx, c}
		return f(ctx, t)
	})
}

func (tx *Transaction) Get(dr *DocumentRef, m MutableModel) error {
	doc, err := tx.tx.Get(dr.DocumentRef)
	if err != nil {
		return err
	}
	if err := docToModel(m, doc); err != nil {
		return err
	}

	// TODO: make expansion optional
	if err := tx.expandModel(m); err != nil {
		return err
	}
	// TODO: configurable retry-loops
	return nil
}

func (tx *Transaction) Documents(q Queryer) *DocumentIterator {
	return &DocumentIterator{tx: tx, it: tx.tx.Documents(q.query().q)}
}

func (tx *Transaction) Set(dr *DocumentRef, m ReadableModel) error {
	sm, err := modelToDoc(m)
	if err != nil {
		return err
	}
	// TODO: transactionally store model history
	return tx.tx.Set(dr.DocumentRef, sm)
}

func (tx *Transaction) Delete(dr *DocumentRef) error {
	return tx.tx.Delete(dr.DocumentRef)
}
