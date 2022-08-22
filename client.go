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

import "cloud.google.com/go/firestore"

// A Client provides access to Firestore via the Calcifer ODM.
type Client struct {
	fs *firestore.Client
}

// NewClient creates a new Calcifier client that uses the given Firestore client.
func NewClient(fs *firestore.Client) *Client {
	return &Client{fs: fs}
}

func (c *Client) Collection(path string) *CollectionRef {
	return &CollectionRef{
		cref: c.fs.Collection(path),
		cli:  c,
		Query: Query{
			cli: c,
			q:   c.fs.Collection(path).Query,
		},
	}
}
