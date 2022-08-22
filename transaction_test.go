package calcifer

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
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
