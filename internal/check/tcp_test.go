package check

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestTCPCheck_Reachable(t *testing.T) {
	laddr := net.TCPAddr{
		IP:   []byte{127, 0, 0, 1},
		Port: 0,
	}

	server, err := net.ListenTCP("tcp", &laddr)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	port := server.Addr().(*net.TCPAddr).Port

	test := TCP{
		Port:    port,
		Timeout: time.Second,
	}

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
	laddr := net.TCPAddr{
		IP:   []byte{127, 0, 0, 1},
		Port: 0,
	}

	server, err := net.ListenTCP("tcp", &laddr)
	if err != nil {
		t.Fatal(err)
	}
	port := server.Addr().(*net.TCPAddr).Port
	if err := server.Close(); err != nil {
		t.Fatal(err)
	}

	test := TCP{
		Port:    port,
		Timeout: time.Second,
	}

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
