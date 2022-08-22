package calcifer

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/iterator"
)

func TestReadModifyWriteTransaction(t *testing.T) {
	ctx := context.Background()
	cli := testClient(t)

	type C struct {
		Model
		N int `calcifer:"n"`
	}
	cs := cli.Collection("c")

	ref := cs.NewDoc()
	err := ref.Set(ctx, C{N: 0})
	assert.NoError(t, err)

	// Run K concurrent transactions that all increment N.
	K := 1 // TODO: the firestore emulator times out if there's concurrency between transactions
	var g errgroup.Group
	for i := 0; i < K; i++ {
		g.Go(func() error {
			return cli.RunTransaction(ctx, func(ctx context.Context, tx *Transaction) error {
				var c C
				if err := tx.Get(ref, &c); err != nil {
					return err
				}
				c.N++
				return tx.Set(ref, c)
			})
		})
	}
	err = g.Wait()
	assert.NoError(t, err)

	var c C
	err = ref.Get(ctx, &c)
	assert.NoError(t, err)
	assert.Equal(t, K, c.N)
}

func TestTransactionalSetAndGetExpansion(t *testing.T) {
	ctx := context.Background()
	cli := testClient(t)

	locationRef := cli.Collection("locations").NewDoc()
	newLocation := Location{
		Name: "Bag End, Hobbiton, The Shire",
	}
	eventRef := cli.Collection("events").NewDoc()
	newEvent := Event{
		Description: "An Unexpected Party",
		Start:       time.Date(1937, time.September, 21, 17, 0, 0, 0, time.UTC),
		End:         time.Date(1937, time.September, 22, 06, 0, 0, 0, time.UTC),
		Location:    &Location{Model: Model{ID: locationRef.ID}},
	}

	err := cli.RunTransaction(ctx, func(ctx context.Context, tx *Transaction) error {
		if err := tx.Set(locationRef, newLocation); err != nil {
			return err
		}
		if err := tx.Set(eventRef, newEvent); err != nil {
			return err
		}
		return nil
	})
	assert.NoError(t, err)

	var savedEvent Event
	err = cli.RunTransaction(ctx, func(ctx context.Context, tx *Transaction) error {
		return tx.Get(eventRef, &savedEvent)
	})
	assert.NoError(t, err)

	assert.Equal(t, eventRef.ID, savedEvent.ID)
	assert.NotZero(t, savedEvent.CreateTime)
	assert.Equal(t, savedEvent.CreateTime, savedEvent.UpdateTime)
	assert.Equal(t, newEvent.Description, savedEvent.Description)
	assert.Equal(t, newEvent.Start, savedEvent.Start)
	assert.Equal(t, newEvent.Location.ID, savedEvent.Location.ID)
	assert.Equal(t, newLocation.Name, savedEvent.Location.Name)
}

func TestTransactionalQueries(t *testing.T) {
	ctx := context.Background()
	cli := testClient(t)

	type C struct {
		Model
		N int `calcifer:"n"`
	}
	cs := cli.Collection("c")

	for i := 0; i < 10; i++ {
		err := cs.NewDoc().Set(ctx, C{N: i})
		assert.NoError(t, err)
	}

	q := cs.Where("n", ">", 2).OrderBy("n", firestore.Asc).Limit(3)
	var ns []int
	err := cli.RunTransaction(ctx, func(ctx context.Context, tx *Transaction) error {
		iter := tx.Documents(q)
		tns := make([]int, 0)
		for {
			var ci C
			err := iter.Next(ctx, &ci)
			if err == iterator.Done {
				break
			}
			assert.NoError(t, err)
			tns = append(tns, ci.N)
		}
		ns = tns
		return nil
	})
	assert.NoError(t, err)

	assert.Equal(t, []int{3, 4, 5}, ns)
}
