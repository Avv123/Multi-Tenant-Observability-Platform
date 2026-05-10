package requests

import "time"

type Filters struct {
	ServiceID   string
	Environment string
	Severity    string
	MetricName  string
	Search      string
	TraceID     string
	StartTime   time.Time
	EndTime     time.Time
	Limit       int
	Offset      int
}
