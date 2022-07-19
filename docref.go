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

type DocumentRef[M Model] struct {
	*firestore.DocumentRef
	cli *Client
}

func (d *DocumentRef[M]) Collection(id string) *CollectionRef[M] {
	return &CollectionRef[M]{
		cref: d.DocumentRef.Collection(id),
		cli:  d.cli,
	}
}

// Get fetches the document referred to by d from Firestore, and unmarshals it into p.
func (d *DocumentRef[M]) Get(ctx context.Context) (*M, error) {
	doc, err := d.DocumentRef.Get(ctx)
	if err != nil {
		return nil, err
	}
	p, err := docToModel[M](doc)
	if err != nil {
		return nil, err
	}

	// TODO: make expansion optional
	if err := expandModel(ctx, d.cli, p); err != nil {
		return nil, err
	}
	// TODO: configurable retry-loops
	return p, nil
}

// Set writes a Model to Firestore at the path referred to by d.
func (d *DocumentRef[M]) Set(ctx context.Context, m M) error {
	sm, err := modelToDoc(m)
	if err != nil {
		return err
	}
	// TODO: transactionally store model history
	_, err = d.DocumentRef.Set(ctx, sm)
	return err
}
