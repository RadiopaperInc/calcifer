package calcifer

import (
	"context"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/assert"
)

func testClient(t *testing.T) *Client {
	ctx := context.Background()
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("Test depends on the firestore emulator")
	}
	cli, err := firestore.NewClient(ctx, "test")
	assert.NoError(t, err)
	return NewClient(cli)
}

type User struct {
	Model

	Email string
}

type Location struct {
	Model

	Name     string
	Capacity int
}

type Event struct {
	Model

	Description string
	Attendees   []User    `calcifer:"attendees,omitempty"`
	Location    *Location `calcifer:"location,ref:locations"`
	Start       time.Time `calcifer:"start"`
	End         time.Time `calcifer:"end"`
}

func TestSetWithNilForeignKey(t *testing.T) {
	ctx := context.Background()
	cli := testClient(t)

	eventRef := cli.Collection("events").NewDoc()
	newEvent := Event{
		Description: "An Unexpected Party",
		Start:       time.Date(1937, time.September, 21, 17, 0, 0, 0, time.UTC),
		End:         time.Date(1937, time.September, 22, 06, 0, 0, 0, time.UTC),
		Location:    nil,
	}
	err := eventRef.Set(ctx, newEvent)
	assert.NoError(t, err)

	// TODO: assert raw firestore contents
}

func TestSetAndGetByID(t *testing.T) {
	ctx := context.Background()
	cli := testClient(t)

	locationRef := cli.Collection("locations").NewDoc()
	newLocation := Location{
		Name: "Bag End, Hobbiton, The Shire",
	}
	err := locationRef.Set(ctx, newLocation)
	assert.NoError(t, err)

	eventRef := cli.Collection("events").NewDoc()
	newEvent := Event{
		Description: "An Unexpected Party",
		Start:       time.Date(1937, time.September, 21, 17, 0, 0, 0, time.UTC),
		End:         time.Date(1937, time.September, 22, 06, 0, 0, 0, time.UTC),
		Location:    &Location{Model: Model{ID: locationRef.ID}},
	}
	err = eventRef.Set(ctx, newEvent)
	assert.NoError(t, err)

	var savedEvent Event
	err = eventRef.Get(ctx, &savedEvent)
	assert.NoError(t, err)

	assert.Equal(t, eventRef.ID, savedEvent.ID)
	assert.NotZero(t, savedEvent.CreateTime)
	assert.Equal(t, savedEvent.CreateTime, savedEvent.UpdateTime)
	assert.Equal(t, newEvent.Description, savedEvent.Description)
	assert.Equal(t, newEvent.Start, savedEvent.Start)
	assert.Equal(t, newEvent.Location.ID, savedEvent.Location.ID)
	assert.Equal(t, newLocation.Name, savedEvent.Location.Name)
}
