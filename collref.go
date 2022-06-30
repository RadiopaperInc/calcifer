package calcifer

import "cloud.google.com/go/firestore"

type CollectionRef firestore.CollectionRef

func (c *CollectionRef) Doc(id string) *DocumentRef {
	return (*DocumentRef)((*firestore.CollectionRef)(c).Doc(id))
}
