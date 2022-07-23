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

	"cloud.google.com/go/firestore"
)

type DocumentRef struct {
	*firestore.DocumentRef
	cli *Client
}

func (d *DocumentRef) Collection(id string) *CollectionRef {
	return &CollectionRef{
		cref: d.DocumentRef.Collection(id),
		cli:  d.cli,
	}
}

// Get fetches the document referred to by d from Firestore, and unmarshals it into p.
func (d *DocumentRef) Get(ctx context.Context, p MutableModel) error {
	doc, err := d.DocumentRef.Get(ctx)
	if err != nil {
		return err
	}
	if err := docToModel(p, doc); err != nil {
		return err
	}

	// TODO: make expansion optional
	if err := d.cli.expandModel(ctx, p); err != nil {
		return err
	}
	// TODO: configurable retry-loops
	return nil
}

// Set writes a Model to Firestore at the path referred to by d.
func (d *DocumentRef) Set(ctx context.Context, m ReadableModel) error {
	sm, err := modelToDoc(m)
	if err != nil {
		return err
	}
	// TODO: transactionally store model history
	_, err = d.DocumentRef.Set(ctx, sm)
	return err
}

// Delete removes from Firestore the document at the path referred to by d if it exists.
func (d *DocumentRef) Delete(ctx context.Context) error {
	// TODO: transactionally store model history
	_, err := d.DocumentRef.Delete(ctx)
	return err
}
