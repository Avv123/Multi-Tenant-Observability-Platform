package responses

import "time"

type LogRow struct {
	EventID     string    `json:"event_id"`
	ServiceName string    `json:"service_name"`
	Environment string    `json:"environment"`
	Severity    string    `json:"severity"`
	Message     string    `json:"message"`
	TraceID     string    `json:"trace_id"`
	OccurredAt  time.Time `json:"occurred_at"`
}

type MetricRow struct {
	EventID     string    `json:"event_id"`
	ServiceName string    `json:"service_name"`
	Environment string    `json:"environment"`
	MetricName  string    `json:"metric_name"`
	Value       float64   `json:"value"`
	OccurredAt  time.Time `json:"occurred_at"`
}

type TraceRow struct {
	TraceID     string    `json:"trace_id"`
	ServiceName string    `json:"service_name"`
	Environment string    `json:"environment"`
	SpanCount   int64     `json:"span_count"`
	FirstSeenAt time.Time `json:"first_seen_at"`
	LastSeenAt  time.Time `json:"last_seen_at"`
}

type TraceSpanRow struct {
	EventID      string    `json:"event_id"`
	TraceID      string    `json:"trace_id"`
	SpanID       string    `json:"span_id"`
	ParentSpanID string    `json:"parent_span_id"`
	Operation    string    `json:"operation"`
	Status       string    `json:"status"`
	ServiceName  string    `json:"service_name"`
	Environment  string    `json:"environment"`
	OccurredAt   time.Time `json:"occurred_at"`
}

type UsageRow struct {
	ServiceID  string `json:"service_id"`
	SignalType string `json:"signal_type"`
	UsageDate  string `json:"usage_date"`
	EventCount int64  `json:"event_count"`
}

type ServiceHealthRow struct {
	ServiceID      string    `json:"service_id"`
	ServiceName    string    `json:"service_name"`
	Environment    string    `json:"environment"`
	LastEventAt    time.Time `json:"last_event_at"`
	EventCount     int64     `json:"event_count"`
	ErrorLogCount  int64     `json:"error_log_count"`
	CriticalCount  int64     `json:"critical_log_count"`
	LatestMetricAt time.Time `json:"latest_metric_at"`
	LatestTraceAt  time.Time `json:"latest_trace_at"`
	HealthStatus   string    `json:"health_status"`
}

type LogSeverityRollupRow struct {
	ServiceID   string    `json:"service_id"`
	ServiceName string    `json:"service_name"`
	Environment string    `json:"environment"`
	Severity    string    `json:"severity"`
	BucketStart time.Time `json:"bucket_start"`
	EventCount  int64     `json:"event_count"`
}

type MetricSeriesRow struct {
	ServiceID    string    `json:"service_id"`
	Environment  string    `json:"environment"`
	MetricName   string    `json:"metric_name"`
	BucketStart  time.Time `json:"bucket_start"`
	SampleCount  int64     `json:"sample_count"`
	SumValue     float64   `json:"sum_value"`
	MinValue     float64   `json:"min_value"`
	MaxValue     float64   `json:"max_value"`
	LastValue    float64   `json:"last_value"`
	AverageValue float64   `json:"average_value"`
}

type TraceLatencyRollupRow struct {
	ServiceID        string    `json:"service_id"`
	ServiceName      string    `json:"service_name"`
	Environment      string    `json:"environment"`
	Operation        string    `json:"operation"`
	BucketStart      time.Time `json:"bucket_start"`
	SpanCount        int64     `json:"span_count"`
	ErrorCount       int64     `json:"error_count"`
	TotalDurationMS  int64     `json:"total_duration_ms"`
	MaxDurationMS    int64     `json:"max_duration_ms"`
	AverageDurationM float64   `json:"average_duration_ms"`
}
