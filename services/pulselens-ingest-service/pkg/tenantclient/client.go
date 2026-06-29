package tenantclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	pulsetenant "github.com/Avv123/pulselens-common/tenant"
	"github.com/Avv123/pulselens-platform/netutil"
)

type Client struct {
	baseURL       string
	internalToken string
	httpClient    *http.Client
}

// New creates a tenant-service HTTP client.
// B6: timeout is 500ms — the ingest hot path cannot afford to wait 10s for a
// slow tenant-service. Callers get a clear 503 quickly rather than a goroutine
// pile-up waiting on a blocked socket.
func New(baseURL, internalToken string) *Client {
	return &Client{
		baseURL:       strings.TrimRight(netutil.NormalizeURL(baseURL), "/"),
		internalToken: internalToken,
		httpClient: &http.Client{
			Timeout: 500 * time.Millisecond,
		},
	}
}

type successResponse struct {
	IsSuccess bool                       `json:"is_success"`
	Data      pulsetenant.ResolvedAPIKey `json:"data"`
}

func (c *Client) ResolveAPIKey(ctx context.Context, apiKey string) (pulsetenant.ResolvedAPIKey, error) {
	requestBody, err := json.Marshal(pulsetenant.ResolveAPIKeyRequest{APIKey: apiKey})
	if err != nil {
		return pulsetenant.ResolvedAPIKey{}, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/api/v1/auth/resolve-api-key", bytes.NewReader(requestBody))
	if err != nil {
		return pulsetenant.ResolvedAPIKey{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Internal-Token", c.internalToken)

	// B9: propagate the incoming request ID across service boundaries.
	if reqID := ctx.Value("request_id"); reqID != nil {
		request.Header.Set("X-Request-Id", fmt.Sprintf("%v", reqID))
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		// Distinguish timeout from other errors for clear observability.
		if strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "timeout") {
			return pulsetenant.ResolvedAPIKey{}, fmt.Errorf("tenant-service unavailable: request timed out after 500ms")
		}
		return pulsetenant.ResolvedAPIKey{}, fmt.Errorf("tenant-service unreachable: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		return pulsetenant.ResolvedAPIKey{}, fmt.Errorf("tenant-service resolve api key returned status %d", response.StatusCode)
	}

	var parsed successResponse
	if err = json.NewDecoder(response.Body).Decode(&parsed); err != nil {
		return pulsetenant.ResolvedAPIKey{}, err
	}

	return parsed.Data, nil
}

var client *Client

func Set(sharedClient *Client) {
	client = sharedClient
}

func Get() *Client {
	return client
}
