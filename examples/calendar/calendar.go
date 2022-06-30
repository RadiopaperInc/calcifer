package main

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/RadiopaperInc/calcifer"
)

type Meeting struct {
	calcifer.Model

	Location string
	Start    time.Time
	End      time.Time
}

func main() {
	ctx := context.Background()
	fsCli, err := firestore.NewClient(ctx, "test")
	if err != nil {
		panic(err)
	}
	cli := calcifer.NewClient(fsCli)
	ref := cli.Collection("meetings").NewDoc()
	newMeeting := Meeting{
		Location: "Bag End, Hobbiton, The Shire",
		Start:    time.Date(1937, time.September, 21, 17, 0, 0, 0, time.UTC),
		End:      time.Date(1937, time.September, 22, 06, 0, 0, 0, time.UTC),
	}

	if err := ref.Set(ctx, newMeeting); err != nil {
		panic(err)
	}

	var storedMeeting Meeting
	if err := ref.Get(ctx, &storedMeeting); err != nil {
		panic(err)
	}
	fmt.Println(ref.ID, storedMeeting.ID)
	fmt.Println(storedMeeting.CreateTime, storedMeeting.UpdateTime)
}
