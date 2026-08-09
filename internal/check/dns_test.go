package check

import (
	"context"
	"testing"
	"time"
)

func TestDNSCheck_Localhost(t *testing.T) {
	test := DNS{
		Timeout: time.Second,
	}

	got, err := test.Check(context.Background(), "localhost")
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !got.Healthy {
		t.Error("Check() Healthy = false, want true")
	}
	if len(got.Response) == 0 {
		t.Error("Check() Response is empty, want at least one address")
	}
}

func TestDNSCheck_InvalidHost(t *testing.T) {
	test := DNS{
		Timeout: time.Second,
	}

	got, err := test.Check(context.Background(), "invalid host")
	if err == nil {
		t.Fatal("Check() error = nil, want invalid host error")
	}
	if got.Healthy {
		t.Error("Check() Healthy = true, want false")
	}
}
