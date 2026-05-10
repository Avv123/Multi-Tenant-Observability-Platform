package archive

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	pulsetelemetry "github.com/omniful/pulselens-common/telemetry"
	platformobjectstore "github.com/omniful/pulselens-platform/objectstore"
)

type Writer struct {
	store *platformobjectstore.Client
	mu    sync.Mutex
}

type ObjectLocation struct {
	Bucket string
	Key    string
	URI    string
}

var writer *Writer

func Set(store *platformobjectstore.Client) {
	writer = &Writer{store: store}
}

func Get() *Writer {
	return writer
}

func (w *Writer) Archive(ctx context.Context, envelope pulsetelemetry.Envelope) (ObjectLocation, error) {
	if w == nil || w.store == nil || !w.store.Enabled() {
		return ObjectLocation{}, nil
	}

	occurredAt := envelope.OccurredAt.UTC()
	key := fmt.Sprintf(
		"%s/%s/%s/%s/%s/%s-%s.json",
		envelope.TenantID,
		occurredAt.Format("2006"),
		occurredAt.Format("01"),
		occurredAt.Format("02"),
		occurredAt.Format("15"),
		envelope.EventType,
		envelope.EventID,
	)
	payload, err := json.Marshal(envelope)
	if err != nil {
		return ObjectLocation{}, err
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	storedKey, err := w.store.PutObject(ctx, key, payload, "application/json")
	if err != nil {
		return ObjectLocation{}, err
	}
	return ObjectLocation{
		Bucket: w.store.Bucket(),
		Key:    storedKey,
		URI:    w.store.URI(storedKey),
	}, nil
}

func (w *Writer) Delete(ctx context.Context, bucket string, key string) error {
	if w == nil || w.store == nil || !w.store.Enabled() {
		return nil
	}
	return w.store.DeleteObject(ctx, bucket, key)
}
