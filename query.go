package calcifer

import (
	"context"

	"cloud.google.com/go/firestore"
)

// Query represents a Firestore query.
//
// Query values are immutable. Each Query method creates
// a new Query; it does not modify the old.
type Query struct {
	query firestore.Query
}

// Where returns a new Query that filters the set of results.
// A Query can have multiple filters.
// The path argument can be a single field or a dot-separated sequence of
// fields, and must not contain any of the runes "˜*/[]".
// The op argument must be one of "==", "!=", "<", "<=", ">", ">=",
// "array-contains", "array-contains-any", "in" or "not-in".
func (q Query) Where(path, op string, value interface{}) Query {
	return Query{query: q.query.Where(path, op, value)}
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
	return Query{q.query.OrderBy(path, dir)}
}

// Limit returns a new Query that specifies the maximum number of first results
// to return. It must not be negative.
func (q Query) Limit(n int) Query {
	return Query{q.query.Limit(n)}
}

// LimitToLast returns a new Query that specifies the maximum number of last
// results to return. It must not be negative.
func (q Query) LimitToLast(n int) Query {
	return Query{q.query.LimitToLast(n)}
}

// TODO: StartAt, StartAfter, EndAt, EndBefore; which require un-wrapping DocumentRefs.

// TODO: Serialize, Deserialize

// Documents returns an iterator over the query's resulting documents.
func (q Query) Documents(ctx context.Context) *DocumentIterator {
	return &DocumentIterator{q.query.Documents(ctx)}
}

type DocumentIterator struct {
	it *firestore.DocumentIterator
}

// Next fetches the next result from Firestore, and unmarshals it into p.
// If error is iterator.Done, no result is unmarshalled. Once Next returns Done,
// all subsequent calls will return
// Done.
func (it *DocumentIterator) Next(p MutableModel) error {
	doc, err := it.it.Next()
	if err != nil {
		return err
	}
	if err := docToModel(p, doc); err != nil {
		return err
	}
	// TODO: expansion
	return nil
}
