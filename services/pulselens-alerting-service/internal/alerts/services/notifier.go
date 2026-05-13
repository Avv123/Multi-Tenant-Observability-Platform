package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	alertmodels "github.com/omniful/pulselens-alerting-service/internal/alerts/models"
	"github.com/omniful/pulselens-platform/config"
	"github.com/omniful/pulselens-platform/netutil"
)

type webhookChannelConfig struct {
	URL            string            `json:"url"`
	Method         string            `json:"method"`
	Headers        map[string]string `json:"headers"`
	TimeoutSeconds int               `json:"timeout_seconds"`
}

type emailChannelConfig struct {
	To            []string `json:"to"`
	From          string   `json:"from"`
	Subject       string   `json:"subject"`
	SubjectPrefix string   `json:"subject_prefix"`
}

func deliverNotification(ctx context.Context, channel alertmodels.NotificationChannel, payload []byte) (string, string, *time.Time) {
	return deliverNotificationWithPolicy(ctx, channel, payload, 3, 200)
}

func deliverNotificationWithPolicy(ctx context.Context, channel alertmodels.NotificationChannel, payload []byte, maxAttempts int, backoffMillis int) (string, string, *time.Time) {
	switch strings.ToLower(strings.TrimSpace(channel.Type)) {
	case "webhook":
		deliveredAt, response, err := deliverWebhook(ctx, channel, payload, maxAttempts, backoffMillis)
		if err != nil {
			return "failed", response, nil
		}
		return "delivered", response, &deliveredAt
	case "slack_webhook":
		deliveredAt, response, err := deliverSlackWebhook(ctx, channel, payload, maxAttempts, backoffMillis)
		if err != nil {
			return "failed", response, nil
		}
		return "delivered", response, &deliveredAt
	case "email":
		deliveredAt, response, err := deliverEmail(channel, payload)
		if err != nil {
			return "failed", response, nil
		}
		return "delivered", response, &deliveredAt
	default:
		now := time.Now().UTC()
		return "delivered", "local-log-delivery", &now
	}
}

func deliverWebhook(ctx context.Context, channel alertmodels.NotificationChannel, payload []byte, maxAttempts int, backoffMillis int) (time.Time, string, error) {
	var cfg webhookChannelConfig
	if err := json.Unmarshal([]byte(channel.Config), &cfg); err != nil {
		return time.Time{}, err.Error(), err
	}
	if strings.TrimSpace(cfg.URL) == "" {
		return time.Time{}, "webhook url is required", fmt.Errorf("webhook url is required")
	}
	cfg.URL = netutil.NormalizeURL(cfg.URL)

	method := strings.ToUpper(strings.TrimSpace(cfg.Method))
	if method == "" {
		method = http.MethodPost
	}
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	client := &http.Client{Timeout: timeout}
	var lastErr error
	var lastResponse string
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	if backoffMillis <= 0 {
		backoffMillis = 200
	}
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		request, err := http.NewRequestWithContext(ctx, method, cfg.URL, bytes.NewReader(payload))
		if err != nil {
			return time.Time{}, err.Error(), err
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-PulseLens-Channel", channel.ID)
		request.Header.Set("X-PulseLens-Attempt", fmt.Sprintf("%d", attempt))
		for key, value := range cfg.Headers {
			request.Header.Set(key, value)
		}

		response, err := client.Do(request)
		if err != nil {
			lastErr = err
			lastResponse = err.Error()
		} else {
			body, _ := io.ReadAll(io.LimitReader(response.Body, 2048))
			_ = response.Body.Close()
			lastResponse = strings.TrimSpace(string(body))
			if lastResponse == "" {
				lastResponse = response.Status
			}
			if response.StatusCode < 300 {
				return time.Now().UTC(), lastResponse, nil
			}
			lastErr = fmt.Errorf("webhook returned status %d", response.StatusCode)
		}

		if attempt < maxAttempts {
			select {
			case <-ctx.Done():
				return time.Time{}, lastResponse, ctx.Err()
			case <-time.After(time.Duration(attempt*backoffMillis) * time.Millisecond):
			}
		}
	}
	return time.Time{}, lastResponse, lastErr
}

func deliverSlackWebhook(ctx context.Context, channel alertmodels.NotificationChannel, payload []byte, maxAttempts int, backoffMillis int) (time.Time, string, error) {
	var incidentPayload map[string]interface{}
	if err := json.Unmarshal(payload, &incidentPayload); err != nil {
		return time.Time{}, err.Error(), err
	}
	slackPayload := map[string]any{
		"text": fmt.Sprintf("%s: %s", notifierStringValue(incidentPayload["event_type"]), notifierStringValue(incidentPayload["title"])),
		"blocks": []map[string]any{
			{
				"type": "section",
				"text": map[string]string{
					"type": "mrkdwn",
					"text": fmt.Sprintf("*%s*\n%s", notifierStringValue(incidentPayload["title"]), notifierStringValue(incidentPayload["summary"])),
				},
			},
			{
				"type": "context",
				"elements": []map[string]string{
					{"type": "mrkdwn", "text": fmt.Sprintf("status: `%s`", notifierStringValue(incidentPayload["status"]))},
					{"type": "mrkdwn", "text": fmt.Sprintf("event: `%s`", notifierStringValue(incidentPayload["event_type"]))},
				},
			},
		},
	}
	payloadBytes, err := json.Marshal(slackPayload)
	if err != nil {
		return time.Time{}, err.Error(), err
	}

	webhookConfig := map[string]any{}
	if err = json.Unmarshal([]byte(channel.Config), &webhookConfig); err != nil {
		return time.Time{}, err.Error(), err
	}
	webhookConfig["method"] = http.MethodPost
	channel.Config = notifierMarshalJSON(webhookConfig)
	return deliverWebhook(ctx, channel, payloadBytes, maxAttempts, backoffMillis)
}

func deliverEmail(channel alertmodels.NotificationChannel, payload []byte) (time.Time, string, error) {
	var cfg emailChannelConfig
	if err := json.Unmarshal([]byte(channel.Config), &cfg); err != nil {
		return time.Time{}, err.Error(), err
	}

	recipients := make([]string, 0, len(cfg.To))
	for _, recipient := range cfg.To {
		if strings.TrimSpace(recipient) != "" {
			recipients = append(recipients, strings.TrimSpace(recipient))
		}
	}
	if len(recipients) == 0 {
		return time.Time{}, "email recipients are required", fmt.Errorf("email recipients are required")
	}

	from := strings.TrimSpace(cfg.From)
	if from == "" {
		from = strings.TrimSpace(config.GetString("smtp.from"))
	}
	if from == "" {
		from = "pulselens@local"
	}

	subject := strings.TrimSpace(cfg.Subject)
	if subject == "" {
		var incidentPayload map[string]interface{}
		_ = json.Unmarshal(payload, &incidentPayload)
		subject = fmt.Sprintf("%s %s", notifierStringValue(incidentPayload["event_type"]), notifierStringValue(incidentPayload["title"]))
	}
	if prefix := strings.TrimSpace(firstNonEmpty(cfg.SubjectPrefix, config.GetString("smtp.subjectPrefix"))); prefix != "" {
		subject = prefix + " " + subject
	}

	host := strings.TrimSpace(config.GetString("smtp.host"))
	if host == "" {
		host = "127.0.0.1"
	}
	port := config.GetInt("smtp.port")
	if port == 0 {
		port = 1025
	}
	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	message := bytes.NewBuffer(nil)
	message.WriteString("From: " + from + "\r\n")
	message.WriteString("To: " + strings.Join(recipients, ", ") + "\r\n")
	message.WriteString("Subject: " + subject + "\r\n")
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString("Content-Type: application/json; charset=utf-8\r\n")
	message.WriteString("\r\n")
	message.Write(payload)

	if err := smtp.SendMail(address, nil, from, recipients, message.Bytes()); err != nil {
		return time.Time{}, err.Error(), err
	}
	return time.Now().UTC(), fmt.Sprintf("email sent to %d recipient(s)", len(recipients)), nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func notifierMarshalJSON(value interface{}) string {
	payload, _ := json.Marshal(value)
	return string(payload)
}

func notifierStringValue(value interface{}) string {
	if typed, ok := value.(string); ok {
		return typed
	}
	return ""
}
