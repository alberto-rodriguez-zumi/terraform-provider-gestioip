package client

import "testing"

func TestInferIPVersion(t *testing.T) {
	t.Parallel()

	if got := inferIPVersion("192.168.1.0"); got != "v4" {
		t.Fatalf("expected v4, got %q", got)
	}

	if got := inferIPVersion("2001:db8::"); got != "v6" {
		t.Fatalf("expected v6, got %q", got)
	}
}
