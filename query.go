package calcifer

import (
	"context"

	"cloud.google.com/go/firestore"
)

// Query represents a Firestore query.
//
// Query values are immutable. Each Query method creates
// a new Query; it does not modify the old.
type Query[M Model] struct {
	cli   *Client
	query firestore.Query
}

// Where returns a new Query that filters the set of results.
// A Query can have multiple filters.
// The path argument can be a single field or a dot-separated sequence of
// fields, and must not contain any of the runes "˜*/[]".
// The op argument must be one of "==", "!=", "<", "<=", ">", ">=",
// "array-contains", "array-contains-any", "in" or "not-in".
func (q Query[M]) Where(path, op string, value interface{}) Query[M] {
	return Query[M]{cli: q.cli, query: q.query.Where(path, op, value)}
}

// OrderBy returns a new Query that specifies the order in which results are
// returned. A Query can have multiple OrderBy/OrderByPath specifications.
// OrderBy appends the specification to the list of existing ones.
//
// The path argument can be a single field or a dot-separated sequence of
// fields, and must not contain any of the runes "˜*/[]".
//
// To order by document name, use the special field path DocumentID.
func (q Query[M]) OrderBy(path string, dir firestore.Direction) Query[M] {
	return Query[M]{cli: q.cli, query: q.query.OrderBy(path, dir)}
}

// Limit returns a new Query that specifies the maximum number of first results
// to return. It must not be negative.
func (q Query[M]) Limit(n int) Query[M] {
	return Query[M]{cli: q.cli, query: q.query.Limit(n)}
}

// LimitToLast returns a new Query that specifies the maximum number of last
// results to return. It must not be negative.
func (q Query[M]) LimitToLast(n int) Query[M] {
	return Query[M]{cli: q.cli, query: q.query.LimitToLast(n)}
}

// TODO: StartAt, StartAfter, EndAt, EndBefore; which require un-wrapping DocumentRefs.

// TODO: Serialize, Deserialize

// Documents returns an iterator over the query's resulting documents.
func (q Query[M]) Documents(ctx context.Context) *DocumentIterator[M] {
	return &DocumentIterator[M]{
		cli: q.cli,
		it:  q.query.Documents(ctx),
	}
}

type DocumentIterator[M Model] struct {
	cli *Client
	it  *firestore.DocumentIterator
}

// Next fetches the next result from Firestore, and unmarshals it into p.
// If error is iterator.Done, no result is unmarshalled. Once Next returns Done,
// all subsequent calls will return
// Done.
func (it *DocumentIterator[M]) Next(ctx context.Context) (*M, error) {
	doc, err := it.it.Next()
	if err != nil {
		return nil, err
	}
	m, err := docToModel[M](doc)
	if err != nil {
		return nil, err
	}

	// TODO: make expansion optional
	if err := expandModel[M](ctx, it.cli, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (it *DocumentIterator[M]) GetAll(ctx context.Context) ([]*M, error) {
	docs, err := it.it.GetAll()
	if err != nil {
		return nil, err
	}
	ms := make([]*M, len(docs))
	for i := 0; i < len(docs); i++ {
		if m, err := docToModel[M](docs[i]); err != nil {
			return nil, err
		} else {
			ms[i] = m
		}
		// TODO: make expansion optional, and parallelize across instances
		if err := expandModel(ctx, it.cli, ms[i]); err != nil {
			return nil, err
		}
	}
	return ms, nil
}
