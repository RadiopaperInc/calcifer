package calcifer

import (
	"context"

	"cloud.google.com/go/firestore"
)

type DocumentRef struct {
	*firestore.DocumentRef
	cli *Client
}

func (d *DocumentRef) Collection(id string) *CollectionRef {
	return &CollectionRef{
		CollectionRef: d.DocumentRef.Collection(id),
		cli:           d.cli,
	}
}

// Get fetches the document referred to by d from Firestore, and unmarshals it into p.
func (d *DocumentRef) Get(ctx context.Context, p MutableModel) error {
	doc, err := d.DocumentRef.Get(ctx)
	if err != nil {
		return err
	}
	if err := docToModel(p, doc); err != nil {
		return err
	}

	// TODO: make expansion optional
	if err := d.cli.expandModel(ctx, p); err != nil {
		return err
	}
	// TODO: configurable retry-loops
	return nil
}

// Set writes a Model to Firestore at the path referred to by d.
func (d *DocumentRef) Set(ctx context.Context, m ReadableModel) error {
	sm, err := modelToDoc(m)
	if err != nil {
		return err
	}
	// TODO: transactionally store model history
	_, err = d.DocumentRef.Set(ctx, sm)
	return err
}
