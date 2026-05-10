package clickhouse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	enabled    bool
	baseURL    string
	database   string
	username   string
	password   string
	httpClient *http.Client
}

type jsonEnvelope[T any] struct {
	Data []T `json:"data"`
}

func New(enabled bool, baseURL, database, username, password string) *Client {
	return &Client{
		enabled:  enabled && strings.TrimSpace(baseURL) != "",
		baseURL:  strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		database: strings.TrimSpace(database),
		username: strings.TrimSpace(username),
		password: password,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.enabled
}

func (c *Client) Ping(ctx context.Context) error {
	if !c.Enabled() {
		return nil
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/ping", nil)
	if err != nil {
		return err
	}
	c.applyAuth(request)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode >= 300 {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("clickhouse ping failed: %s", strings.TrimSpace(string(body)))
	}
	return nil
}

func (c *Client) Exec(ctx context.Context, query string) error {
	_, err := c.do(ctx, query, nil)
	return err
}

func (c *Client) InsertJSONEachRow(ctx context.Context, table string, payloads [][]byte) error {
	if !c.Enabled() || len(payloads) == 0 {
		return nil
	}

	var body bytes.Buffer
	for _, payload := range payloads {
		body.Write(payload)
		body.WriteByte('\n')
	}
	_, err := c.do(ctx, fmt.Sprintf("INSERT INTO %s FORMAT JSONEachRow", table), &body)
	return err
}

func Select[T any](ctx context.Context, client *Client, query string) ([]T, error) {
	rows := make([]T, 0)
	if client == nil || !client.Enabled() {
		return rows, nil
	}

	responseBody, err := client.do(ctx, ensureJSONFormat(query), nil)
	if err != nil {
		return nil, err
	}

	var envelope jsonEnvelope[T]
	if err = json.Unmarshal(responseBody, &envelope); err != nil {
		return nil, err
	}
	return envelope.Data, nil
}

func (c *Client) do(ctx context.Context, query string, body io.Reader) ([]byte, error) {
	if !c.Enabled() {
		return nil, nil
	}

	requestURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}

	queryValues := requestURL.Query()
	if c.database != "" && !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(query)), "CREATE DATABASE") {
		queryValues.Set("database", c.database)
	}
	queryValues.Set("output_format_json_quote_64bit_integers", "0")
	queryValues.Set("query", query)
	requestURL.RawQuery = queryValues.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL.String(), body)
	if err != nil {
		return nil, err
	}
	c.applyAuth(request)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode >= 300 {
		return nil, fmt.Errorf("clickhouse query failed (%d): %s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	return responseBody, nil
}

func (c *Client) applyAuth(request *http.Request) {
	if c.username != "" {
		request.SetBasicAuth(c.username, c.password)
	}
}

func ensureJSONFormat(query string) string {
	upper := strings.ToUpper(query)
	if strings.Contains(upper, " FORMAT ") {
		return query
	}
	return query + " FORMAT JSON"
}
