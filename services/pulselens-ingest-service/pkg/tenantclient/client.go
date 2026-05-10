package tenantclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	pulsetenant "github.com/omniful/pulselens-common/tenant"
)

type Client struct {
	baseURL       string
	internalToken string
	httpClient    *http.Client
}

func New(baseURL, internalToken string) *Client {
	return &Client{
		baseURL:       strings.TrimRight(baseURL, "/"),
		internalToken: internalToken,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
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

	response, err := c.httpClient.Do(request)
	if err != nil {
		return pulsetenant.ResolvedAPIKey{}, err
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
