package requests

type CreateAlertPolicyRequest struct {
	Name                      string   `json:"name"`
	Description               string   `json:"description"`
	MaxDeliveryAttempts       int      `json:"max_delivery_attempts"`
	DeliveryBackoffMillis     int      `json:"delivery_backoff_millis"`
	EscalationIntervalMinutes int      `json:"escalation_interval_minutes"`
	MaxEscalations            int      `json:"max_escalations"`
	RepeatNotificationMinutes int      `json:"repeat_notification_minutes"`
	OpenChannelTypes          []string `json:"open_channel_types"`
	AckChannelTypes           []string `json:"ack_channel_types"`
	ResolveChannelTypes       []string `json:"resolve_channel_types"`
	EscalationChannelTypes    []string `json:"escalation_channel_types"`
	Active                    *bool    `json:"active"`
}

type UpdateAlertPolicyRequest struct {
	Name                      string   `json:"name"`
	Description               string   `json:"description"`
	MaxDeliveryAttempts       int      `json:"max_delivery_attempts"`
	DeliveryBackoffMillis     int      `json:"delivery_backoff_millis"`
	EscalationIntervalMinutes int      `json:"escalation_interval_minutes"`
	MaxEscalations            int      `json:"max_escalations"`
	RepeatNotificationMinutes int      `json:"repeat_notification_minutes"`
	OpenChannelTypes          []string `json:"open_channel_types"`
	AckChannelTypes           []string `json:"ack_channel_types"`
	ResolveChannelTypes       []string `json:"resolve_channel_types"`
	EscalationChannelTypes    []string `json:"escalation_channel_types"`
	Active                    *bool    `json:"active"`
}
