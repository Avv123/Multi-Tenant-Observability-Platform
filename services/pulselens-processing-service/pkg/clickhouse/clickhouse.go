package clickhouse

import platformclickhouse "github.com/omniful/pulselens-platform/clickhouse"

var client *platformclickhouse.Client

func Set(value *platformclickhouse.Client) {
	client = value
}

func Get() *platformclickhouse.Client {
	return client
}
