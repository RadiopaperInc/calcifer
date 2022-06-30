package calcifer

import (
	"crypto/rand"
	"fmt"

	"cloud.google.com/go/firestore"
)

type CollectionRef firestore.CollectionRef

func (c *CollectionRef) Doc(id string) *DocumentRef {
	return (*DocumentRef)((*firestore.CollectionRef)(c).Doc(id))
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
