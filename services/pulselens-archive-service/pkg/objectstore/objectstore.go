package objectstore

import platformobjectstore "github.com/omniful/pulselens-platform/objectstore"

var client *platformobjectstore.Client

func Set(store *platformobjectstore.Client) {
	client = store
}

func Get() *platformobjectstore.Client {
	return client
}
