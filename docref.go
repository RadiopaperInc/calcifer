package calcifer

import (
	"context"

	"cloud.google.com/go/firestore"
)

type DocumentRef firestore.DocumentRef

func (d *DocumentRef) Collection(id string) *CollectionRef {
	return (*CollectionRef)((*firestore.DocumentRef)(d).Collection(id))
}

// Get fetches the document referred to by d from Firestore, and unmarshals it into p.
func (d *DocumentRef) Get(ctx context.Context, p MutableModel) error {
	fd := (*firestore.DocumentRef)(d)
	doc, err := fd.Get(ctx)
	if err != nil {
		return err
	}
	if err := docToModel(p, doc); err != nil {
		return err
	}
	// TODO: optionally fetch foreign-key refs
	// TODO: configurable retry-loops
	return nil
}

func (d *DocumentRef) Set(ctx context.Context, m ReadableModel) error {
	fd := (*firestore.DocumentRef)(d)
	_, err := fd.Set(ctx, m) // TODO: use our own struct tags
	return err
}
