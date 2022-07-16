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
		CollectionRef: c.fs.Collection(path),
		cli:           c,
	}
}
