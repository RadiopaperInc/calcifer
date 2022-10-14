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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/assert"
)

func testClient(t *testing.T) *Client {
	ctx := context.Background()
	eh := os.Getenv("FIRESTORE_EMULATOR_HOST")
	if eh == "" {
		t.Skip("Test depends on the firestore emulator")
	}
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://%s/emulator/v1/projects/test/databases/(default)/documents", eh), nil)
	assert.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	if resp.StatusCode != http.StatusOK {
		errBody, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		t.Errorf("%d: %s", resp.StatusCode, string(errBody))
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

type Beverage struct {
	Model

	Name     string
	Quantity int
}

type Event struct {
	Model

	Description string
	Attendees   []User      `calcifer:"attendees,ref:users,omitempty"`
	Beverages   []*Beverage `calcifer:"beverages,ref:beverages"`
	Location    *Location   `calcifer:"location,ref:locations"`
	Start       time.Time   `calcifer:"start"`
	End         time.Time   `calcifer:"end"`
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

	users := cli.Collection("users")
	bilboRef, gandalfRef, thorinRef := users.NewDoc(), users.NewDoc(), users.NewDoc()
	bilbo, gandalf, thorin := User{Email: "bilbo@theshire.net"}, User{Email: "gandalf@middle-earth.org"}, User{Email: "thorin@underthemountain.com"}
	assert.NoError(t, bilboRef.Set(ctx, bilbo))
	assert.NoError(t, gandalfRef.Set(ctx, gandalf))
	assert.NoError(t, thorinRef.Set(ctx, thorin))
	bilbo.ID = bilboRef.ID
	gandalf.ID = gandalfRef.ID
	thorin.ID = thorinRef.ID

	aleRef := cli.Collection("beverages").NewDoc()
	ale := &Beverage{Name: "ale", Quantity: 14, Model: Model{ID: aleRef.ID}}
	assert.NoError(t, aleRef.Set(ctx, ale))

	locationRef := cli.Collection("locations").NewDoc()
	newLocation := Location{
		Name: "Bag End, Hobbiton, The Shire",
	}
	assert.NoError(t, locationRef.Set(ctx, newLocation))

	eventRef := cli.Collection("events").NewDoc()
	newEvent := Event{
		Description: "An Unexpected Party",
		Start:       time.Date(1937, time.September, 21, 17, 0, 0, 0, time.UTC),
		End:         time.Date(1937, time.September, 22, 06, 0, 0, 0, time.UTC),
		Location:    &Location{Model: Model{ID: locationRef.ID}},
		Attendees:   []User{bilbo, gandalf, thorin},
		Beverages:   []*Beverage{ale},
	}
	assert.NoError(t, eventRef.Set(ctx, newEvent))

	var savedEvent Event
	assert.NoError(t, eventRef.Get(ctx, &savedEvent))

	var zeroTime time.Time // clear timestamps for comparison
	for i := range savedEvent.Attendees {
		savedEvent.Attendees[i].CreateTime = zeroTime
		savedEvent.Attendees[i].UpdateTime = zeroTime
	}

	assert.Equal(t, eventRef.ID, savedEvent.ID)
	assert.NotZero(t, savedEvent.CreateTime)
	assert.Equal(t, savedEvent.CreateTime, savedEvent.UpdateTime)
	assert.Equal(t, "ale", savedEvent.Beverages[0].Name)
	assert.Equal(t, 14, savedEvent.Beverages[0].Quantity)
	assert.Equal(t, []User{bilbo, gandalf, thorin}, savedEvent.Attendees)
	assert.Equal(t, newEvent.Description, savedEvent.Description)
	assert.Equal(t, newEvent.Start, savedEvent.Start)
	assert.Equal(t, newEvent.Location.ID, savedEvent.Location.ID)
	assert.Equal(t, newLocation.Name, savedEvent.Location.Name)
}

func TestFirestoreNestedStructPointers(t *testing.T) {
	ctx := context.Background()
	cli, err := firestore.NewClient(ctx, "test")
	assert.NoError(t, err)

	type C struct {
		X string `firestore:"x"`
	}

	type A struct {
		B []*C `firestore:"b"`
	}

	a1 := A{B: []*C{&C{X: "x"}}}
	_, err = cli.Collection("a").Doc("1").Set(ctx, a1)
	assert.NoError(t, err)

	var a2 A
	doc, err := cli.Collection("a").Doc("1").Get(ctx)
	assert.NoError(t, err)

	err = doc.DataTo(&a2)
	assert.NoError(t, err)

	assert.Equal(t, a1, a2)
}
