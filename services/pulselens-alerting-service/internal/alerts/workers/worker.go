package workers

import (
	"context"
	"fmt"
	"time"

	alertservices "github.com/Avv123/pulselens-alerting-service/internal/alerts/services"
	serviceredis "github.com/Avv123/pulselens-alerting-service/pkg/redis"
	"github.com/Avv123/pulselens-platform/config"
	"github.com/Avv123/pulselens-platform/idgen"
	"github.com/Avv123/pulselens-platform/lock"
	"github.com/Avv123/pulselens-platform/logging"
)

func Run(ctx context.Context) error {
	service := alertservices.New()
	ticker := time.NewTicker(time.Duration(config.GetInt("evaluation.intervalSeconds")) * time.Second)
	defer ticker.Stop()

	owner := idgen.New("alert-worker")
	mutex := lock.NewRedisLock(serviceredis.Get())
	lockKey := "pulselens:alerting:evaluator"
	lockTTL := time.Duration(config.GetInt("evaluation.lockTTLSeconds")) * time.Second
	if lockTTL <= 0 {
		lockTTL = 25 * time.Second
	}

	for {
		if err := evaluateOnce(ctx, service, mutex, lockKey, owner, lockTTL); err != nil {
			logging.Errorf("alert evaluation failed: %v", err)
		}

		select {
		case <-ctx.Done():
			_, _ = mutex.ReleaseOwner(context.Background(), lockKey, owner)
			return nil
		case <-ticker.C:
		}
	}
}

func evaluateOnce(ctx context.Context, service *alertservices.Service, mutex *lock.RedisLock, lockKey string, owner string, ttl time.Duration) error {
	acquired, err := mutex.Acquire(ctx, lockKey, owner, ttl)
	if err != nil {
		return err
	}
	if !acquired {
		return nil
	}
	defer func() {
		released, releaseErr := mutex.ReleaseOwner(context.Background(), lockKey, owner)
		if releaseErr != nil {
			logging.Errorf("failed to release alert lock: %v", releaseErr)
			return
		}
		if !released {
			logging.Errorf("alert lock owner changed before release key=%s owner=%s", lockKey, owner)
		}
	}()

	if err = service.EvaluateAll(ctx); err != nil {
		return fmt.Errorf("evaluate all rules: %w", err)
	}
	if err = service.EvaluateEscalations(ctx); err != nil {
		return fmt.Errorf("evaluate escalations: %w", err)
	}
	return nil
}
