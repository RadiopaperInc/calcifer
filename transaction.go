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

func (tx *Transaction) GetAll(drs []*DocumentRef, ms any) error {
	return nil
}

func (tx *Transaction) Documents(q Queryer) *DocumentIterator {
	return nil
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
