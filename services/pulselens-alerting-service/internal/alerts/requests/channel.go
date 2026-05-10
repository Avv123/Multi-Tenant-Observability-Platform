package requests

type CreateNotificationChannelRequest struct {
	Name   string      `json:"name"`
	Type   string      `json:"type"`
	Config interface{} `json:"config"`
	Active *bool       `json:"active"`
}
