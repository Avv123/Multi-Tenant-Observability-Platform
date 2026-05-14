package readiness

import (
	"context"
	"time"
)

type Check struct {
	Name string
	Kind string
	Fn   func(context.Context) error
}

func Run(ctx context.Context, timeout time.Duration, checks ...Check) []DependencyStatus {
	rows := make([]DependencyStatus, 0, len(checks))
	for _, check := range checks {
		checkCtx := ctx
		cancel := func() {}
		if timeout > 0 {
			checkCtx, cancel = context.WithTimeout(ctx, timeout)
		}
		err := check.Fn(checkCtx)
		cancel()
		row := DependencyStatus{
			Name:      check.Name,
			Kind:      check.Kind,
			Status:    "healthy",
			CheckedAt: time.Now().UTC().Format(time.RFC3339),
		}
		if err != nil {
			row.Status = "down"
			row.Error = err.Error()
		}
		rows = append(rows, row)
	}
	return rows
}

func AllHealthy(rows []DependencyStatus) bool {
	for _, row := range rows {
		if !Healthy(row) {
			return false
		}
	}
	return true
}
