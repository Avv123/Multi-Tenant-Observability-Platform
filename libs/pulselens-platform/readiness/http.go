package readiness

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Avv123/pulselens-platform/netutil"
)

func CheckHTTP(ctx context.Context, rawURL string, headers map[string]string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, netutil.NormalizeURL(rawURL), nil)
	if err != nil {
		return err
	}
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
		return fmt.Errorf("http readiness failed (%d): %s", response.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}
