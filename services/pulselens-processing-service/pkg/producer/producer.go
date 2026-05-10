package producer

import (
	"sync"

	platformkafka "github.com/omniful/pulselens-platform/kafka"
)

var (
	client *platformkafka.Producer
	once   sync.Once
)

func Set(producer *platformkafka.Producer) {
	once.Do(func() {
		client = producer
	})
}

func Get() *platformkafka.Producer {
	return client
}
