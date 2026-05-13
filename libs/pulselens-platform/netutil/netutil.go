package netutil

import (
	"net"
	"net/url"
	"strings"
)

func NormalizeHostPort(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	host, port, err := net.SplitHostPort(value)
	if err != nil {
		if strings.EqualFold(value, "localhost") {
			return "127.0.0.1"
		}
		return value
	}
	if strings.EqualFold(host, "localhost") {
		host = "127.0.0.1"
	}
	return net.JoinHostPort(host, port)
}

func NormalizeURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return raw
	}
	host := parsed.Hostname()
	if !strings.EqualFold(host, "localhost") {
		return raw
	}
	port := parsed.Port()
	if port == "" {
		parsed.Host = "127.0.0.1"
	} else {
		parsed.Host = net.JoinHostPort("127.0.0.1", port)
	}
	return parsed.String()
}

func NormalizeDSNHost(dsn string) string {
	parts := strings.Fields(strings.TrimSpace(dsn))
	for index, part := range parts {
		if !strings.HasPrefix(strings.ToLower(part), "host=") {
			continue
		}
		host := strings.TrimPrefix(part, "host=")
		if strings.EqualFold(host, "localhost") {
			parts[index] = "host=127.0.0.1"
		}
	}
	return strings.Join(parts, " ")
}

func NormalizeHosts(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, NormalizeHostPort(value))
	}
	return result
}
