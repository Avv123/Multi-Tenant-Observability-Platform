package workers

import (
	"context"

	telemetryservices "github.com/omniful/pulselens-processing-service/internal/telemetry/services"
)

func Run(ctx context.Context) error {
	return telemetryservices.RunWorkers(ctx)
}
