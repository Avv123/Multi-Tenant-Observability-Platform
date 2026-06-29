package readiness

import (
	"context"

	platformkafka "github.com/Avv123/pulselens-platform/kafka"
)

func CheckKafka(ctx context.Context, brokers []string) error {
	return platformkafka.Ping(ctx, brokers)
}
