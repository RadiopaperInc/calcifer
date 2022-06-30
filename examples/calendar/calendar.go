package main

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/RadiopaperInc/calcifer"
)

type Meeting struct {
	calcifer.Model
}

func main() {
	ctx := context.Background()
	fsCli, err := firestore.NewClient(ctx, "test")
	if err != nil {
		panic(err)
	}
	cli := calcifer.NewClient(fsCli)
	ref := cli.Collection("meetings").Doc("a")
	var meeting Meeting
	if err := ref.Get(ctx, &meeting); err != nil {
		panic(err)
	}
	fmt.Println(meeting.ID)
}
