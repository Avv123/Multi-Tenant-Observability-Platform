package requests

type CreateReplayJobRequest struct {
	ServiceID string `json:"service_id"`
	EventType string `json:"event_type"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}
