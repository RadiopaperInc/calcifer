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
	cs := cli.Collection("c")

	for i := 0; i < 10; i++ {
		err := cs.NewDoc().Set(ctx, C{N: i})
		assert.NoError(t, err)
	}

	q := cs.Where("n", ">", 2).OrderBy("n", firestore.Asc).Limit(3)
	iter := q.Documents(ctx)
	ns := make([]int, 0)
	for {
		var ci C
		err := iter.Next(ctx, &ci)
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

func TestQueryGetAll(t *testing.T) {
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

	var p []Post
	err := posts.OrderBy("body", firestore.Desc).Documents(ctx).GetAll(ctx, &p)
	assert.Equal(t, 2, len(p))
	assert.NoError(t, err)
	assert.Equal(t, evan.ID, p[0].Author.ID)
	assert.Equal(t, "Evan", p[0].Author.Name)
	assert.Equal(t, "Dave", p[1].Author.Name)
	assert.Equal(t, dave.ID, p[1].Author.ID)
}

// TODO: test GetAll on Posts that have:
// * nil Author
// * non-nil Author with empty Author.ID
// * non-nil Author with Author.ID for non-existent User
