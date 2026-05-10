package telemetry

import "time"

type EventType string

const (
	EventTypeLog    EventType = "log"
	EventTypeMetric EventType = "metric"
	EventTypeTrace  EventType = "trace"
	EventTypeCustom EventType = "custom"
)

type Envelope struct {
	EventID       string                 `json:"event_id"`
	TenantID      string                 `json:"tenant_id"`
	TenantName    string                 `json:"tenant_name,omitempty"`
	ServiceID     string                 `json:"service_id"`
	ServiceName   string                 `json:"service_name"`
	Environment   string                 `json:"environment"`
	ShardID       int                    `json:"shard_id"`
	EventType     EventType              `json:"event_type"`
	SchemaVersion string                 `json:"schema_version"`
	OccurredAt    time.Time              `json:"occurred_at"`
	ReceivedAt    time.Time              `json:"received_at"`
	TraceID       string                 `json:"trace_id,omitempty"`
	Severity      string                 `json:"severity,omitempty"`
	RetryCount    int                    `json:"retry_count,omitempty"`
	NextAttemptAt time.Time              `json:"next_attempt_at,omitempty"`
	Payload       map[string]interface{} `json:"payload"`
}

type LogPayload struct {
	Message    string                 `json:"message"`
	Level      string                 `json:"level"`
	Logger     string                 `json:"logger,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

type MetricPayload struct {
	MetricName string                 `json:"metric_name"`
	Value      float64                `json:"value"`
	Unit       string                 `json:"unit,omitempty"`
	Labels     map[string]interface{} `json:"labels,omitempty"`
}

type TracePayload struct {
	SpanID       string                 `json:"span_id"`
	ParentSpanID string                 `json:"parent_span_id,omitempty"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      time.Time              `json:"end_time"`
	Status       string                 `json:"status"`
	Operation    string                 `json:"operation"`
	Attributes   map[string]interface{} `json:"attributes,omitempty"`
}

type BatchIngestRequest struct {
	Events []ClientEvent `json:"events" validate:"required,min=1,dive"`
}

type ClientEvent struct {
	EventID       string                 `json:"event_id,omitempty"`
	EventType     EventType              `json:"event_type" validate:"required"`
	SchemaVersion string                 `json:"schema_version"`
	OccurredAt    time.Time              `json:"occurred_at"`
	TraceID       string                 `json:"trace_id,omitempty"`
	Severity      string                 `json:"severity,omitempty"`
	Payload       map[string]interface{} `json:"payload" validate:"required"`
}

type QueryFilters struct {
	TenantID    string
	ServiceID   string
	ServiceName string
	Environment string
	StartTime   time.Time
	EndTime     time.Time
	TraceID     string
	Severity    string
	MetricName  string
	EventType   EventType
	Search      string
	Limit       int
	Offset      int
}
