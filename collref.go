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
	"crypto/rand"
	"fmt"

	"cloud.google.com/go/firestore"
)

type CollectionRef struct {
	cref *firestore.CollectionRef
	cli  *Client

	// Use the methods of Query on a CollectionRef to create and run queries.
	Query
}

func (c *CollectionRef) Doc(id string) *DocumentRef {
	return &DocumentRef{
		DocumentRef: c.cref.Doc(id),
		cli:         c.cli,
	}
}

// NewDoc returns a DocumentRef with a uniquely generated ID.
func (c *CollectionRef) NewDoc() *DocumentRef {
	return c.Doc(uniqueID())
}

const alphanum = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func uniqueID() string {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("calcifer: crypto/rand.Read error: %v", err))
	}
	for i, byt := range b {
		b[i] = alphanum[int(byt)%len(alphanum)]
	}
	return string(b)
}
