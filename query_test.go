package calcifer

import (
	"context"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/iterator"
)

func TestQueryIterator(t *testing.T) {
	ctx := context.Background()
	cli := testClient(t)

	type C struct {
		Model
		N int `calcifer:"n"`
	}
	cs := Collection[C](cli, "c")

	for i := 0; i < 10; i++ {
		err := cs.NewDoc().Set(ctx, C{N: i})
		assert.NoError(t, err)
	}

	q := cs.Where("n", ">", 2).OrderBy("n", firestore.Asc).Limit(3)
	iter := q.Documents(ctx)
	ns := make([]int, 0)
	for {
		ci, err := iter.Next(ctx)
		if err == iterator.Done {
			break
		}
		assert.NoError(t, err)
		ns = append(ns, ci.N)
	}
	assert.Equal(t, []int{3, 4, 5}, ns)
}

func TestQueryIteratorExpansion(t *testing.T) {
	ctx := context.Background()
	cli := testClient(t)

	type User struct {
		Model
		Name string `calcifer:"name"`
	}

	type Post struct {
		Model
		Body   string `calcifer:"body"`
		Author *User  `calcifer:"author,ref:users"`
	}

	users := cli.Collection("users")
	dave := users.NewDoc()
	assert.NoError(t, dave.Set(ctx, User{Name: "Dave"}))
	evan := users.NewDoc()
	assert.NoError(t, evan.Set(ctx, User{Name: "Evan"}))

	posts := cli.Collection("posts")

	assert.NoError(t, posts.NewDoc().Set(ctx, Post{
		Body:   "Hello, World!",
		Author: &User{Model: Model{ID: dave.ID}},
	}))

	assert.NoError(t, posts.NewDoc().Set(ctx, Post{
		Body:   "Time for a long walk",
		Author: &User{Model: Model{ID: evan.ID}},
	}))

	iter := posts.OrderBy("body", firestore.Desc).Documents(ctx)
	var p1, p2 Post
	assert.NoError(t, iter.Next(ctx, &p1))
	assert.NoError(t, iter.Next(ctx, &p2))
	assert.Equal(t, iterator.Done, iter.Next(ctx, nil))

	assert.Equal(t, "Evan", p1.Author.Name)
	assert.Equal(t, "Dave", p2.Author.Name)
}

func TestQueryGetAllExpansion(t *testing.T) {
	ctx := context.Background()
	cli := testClient(t)

	type User struct {
		Model
		Name string `calcifer:"name"`
	}

	type Post struct {
		Model
		Body   string `calcifer:"body"`
		Author *User  `calcifer:"author,ref:users"`
	}

	users := cli.Collection("users")
	dave := users.NewDoc()
	assert.NoError(t, dave.Set(ctx, User{Name: "Dave"}))
	evan := users.NewDoc()
	assert.NoError(t, evan.Set(ctx, User{Name: "Evan"}))

	posts := cli.Collection("posts")

	assert.NoError(t, posts.NewDoc().Set(ctx, Post{
		Body:   "Hello, World!",
		Author: &User{Model: Model{ID: dave.ID}},
	}))

	assert.NoError(t, posts.NewDoc().Set(ctx, Post{
		Body:   "Time for a long walk",
		Author: &User{Model: Model{ID: evan.ID}},
	}))

	var ps []Post
	err := posts.OrderBy("body", firestore.Desc).Documents(ctx).GetAll(ctx, &ps)
	assert.NoError(t, err)

	assert.Equal(t, "Evan", ps[0].Author.Name)
	assert.Equal(t, "Dave", ps[1].Author.Name)
}
