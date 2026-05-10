package idgen

import "testing"

func TestNewGeneratesMonotonicIDs(t *testing.T) {
	Configure(7)

	first := New("evt")
	second := New("evt")

	if first == second {
		t.Fatalf("expected unique ids")
	}
	if first >= second {
		t.Fatalf("expected lexical order to increase, got %s then %s", first, second)
	}
}

func TestNodeIDFromServiceNameIsStable(t *testing.T) {
	first := NodeIDFromServiceName("pulselens-query-service")
	second := NodeIDFromServiceName("pulselens-query-service")

	if first != second {
		t.Fatalf("expected stable node id, got %d and %d", first, second)
	}
}
