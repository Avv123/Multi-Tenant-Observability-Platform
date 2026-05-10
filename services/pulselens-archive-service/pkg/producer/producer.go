package producer

import platformkafka "github.com/omniful/pulselens-platform/kafka"

var client *platformkafka.Producer

func Set(db *platformkafka.Producer) {
	client = db
}

func Get() *platformkafka.Producer {
	return client
}
