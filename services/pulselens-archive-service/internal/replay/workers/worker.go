package workers

import (
	"context"
	"time"

	replayservices "github.com/omniful/pulselens-archive-service/internal/replay/services"
	serviceredis "github.com/omniful/pulselens-archive-service/pkg/redis"
	"github.com/omniful/pulselens-platform/config"
	"github.com/omniful/pulselens-platform/idgen"
	"github.com/omniful/pulselens-platform/lock"
	"github.com/omniful/pulselens-platform/logging"
)

func Run(ctx context.Context) error {
	service := replayservices.New()
	interval := time.Duration(config.GetInt("replay.pollSeconds")) * time.Second
	if interval <= 0 {
		interval = 3 * time.Second
	}
	lockTTL := time.Duration(config.GetInt("replay.lockTTLSeconds")) * time.Second
	if lockTTL <= 0 {
		lockTTL = 20 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	owner := idgen.New("replay-worker")
	mutex := lock.NewRedisLock(serviceredis.Get())

	for {
		acquired, err := mutex.Acquire(ctx, "pulselens:archive:replay-worker", owner, lockTTL)
		if err != nil {
			return err
		}
		if acquired {
			if err = service.RunReplayJobs(ctx); err != nil {
				logging.Errorf("replay worker failed: %v", err)
			}
			_, _ = mutex.ReleaseOwner(context.Background(), "pulselens:archive:replay-worker", owner)
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
