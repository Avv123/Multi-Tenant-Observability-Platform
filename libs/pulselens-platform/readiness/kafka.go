package readiness

import (
	"context"

	platformkafka "github.com/omniful/pulselens-platform/kafka"
)

func CheckKafka(ctx context.Context, brokers []string) error {
	return platformkafka.Ping(ctx, brokers)
}
