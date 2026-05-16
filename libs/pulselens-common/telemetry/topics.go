package telemetry

// TopicForEventType returns the canonical Kafka topic name for a given signal type.
// B16: This function was duplicated verbatim between ingest-service (topicFor) and
// processing-service (retryTargetTopic). A shared implementation here ensures that
// both services always route to the same topic for the same event type, eliminating
// a class of hard-to-debug misrouting bugs when topic names change.
//
// The topic config keys follow the pattern kafka.topics.<signal>.
// Callers are expected to resolve the actual topic name from their config.
// This function returns the config key path, not the resolved value, so it stays
// config-loader-agnostic.
func TopicConfigKey(eventType EventType) string {
	switch eventType {
	case EventTypeMetric:
		return "kafka.topics.metrics"
	case EventTypeTrace:
		return "kafka.topics.traces"
	case EventTypeCustom:
		return "kafka.topics.custom"
	default:
		return "kafka.topics.logs"
	}
}
