package calcifer

import (
	"context"

	"cloud.google.com/go/firestore"
)

type Transaction struct {
	tx *firestore.Transaction
}

type TransactionOption any

func (c *Client) RunTransaction(ctx context.Context, f func(context.Context, *Transaction) error, opts ...TransactionOption) (err error) {
	return c.fs.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		t := &Transaction{tx}
		return f(ctx, t)
	})
}

func (tx *Transaction) Get(dr *DocumentRef, m MutableModel) error {
	return nil
}

func (tx *Transaction) GetAll(drs []*DocumentRef, ms any) error {
	return nil
}

func (tx *Transaction) Documents(q Queryer) *DocumentIterator {
	return nil
}

func (tx *Transaction) Set(dr *DocumentRef, m ReadableModel) error {
	return nil
}

func (tx *Transaction) Delete(dr *DocumentRef) error {
	return tx.tx.Delete(dr.DocumentRef)
}
