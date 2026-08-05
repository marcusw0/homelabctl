package check

import (
	"context"
	"log"
	"net"
	"testing"
	"time"
)

var test = TCP{
	Port:    1234,
	Timeout: time.Second,
}

func TestTCPCheck_Reachable(t *testing.T) {
	laddr := net.TCPAddr{
		IP:   []byte{127, 0, 0, 1},
		Port: 1234,
	}

	server, err := net.ListenTCP("tcp", &laddr)
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	got, err := test.Check(context.Background(), "127.0.0.1")

	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !got.Healthy {
		t.Errorf(
			"Check() Healthy = false, want true; message: %s",
			got.Message,
		)
	}
}

func TestTCPCheck_Unreachable(t *testing.T) {
	got, err := test.Check(context.Background(), "127.0.0.1")

	if err != nil {
		t.Fatalf("Check() error = %v; unreachable is a valid result", err)
	}
	if got.Healthy {
		t.Error("Check() Healthy = true, want false")
	}
	if got.Message == "" {
		t.Error("Check() Message is empty, want diagnostic information")
	}
}
