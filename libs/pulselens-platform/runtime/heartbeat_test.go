package runtime

import (
	"testing"
	"time"
)

func TestStatusFor(t *testing.T) {
	now := time.Now().UTC()

	if status := statusFor(now.Add(-2*time.Second), 10, now); status != "healthy" {
		t.Fatalf("expected healthy, got %s", status)
	}
	if status := statusFor(now.Add(-6*time.Second), 10, now); status != "degraded" {
		t.Fatalf("expected degraded, got %s", status)
	}
	if status := statusFor(now.Add(-12*time.Second), 10, now); status != "down" {
		t.Fatalf("expected down, got %s", status)
	}
}
