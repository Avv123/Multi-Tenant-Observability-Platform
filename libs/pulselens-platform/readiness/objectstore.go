package readiness

import (
	"context"

	platformobjectstore "github.com/Avv123/pulselens-platform/objectstore"
)

func CheckObjectStore(ctx context.Context, enabled bool, endpoint, region, accessKey, secretKey, bucket, prefix string, forcePathStyle bool) error {
	client, err := platformobjectstore.New(enabled, endpoint, region, accessKey, secretKey, bucket, prefix, forcePathStyle)
	if err != nil {
		return err
	}
	return client.Ping(ctx)
}
