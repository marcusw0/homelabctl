package check

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestTLSCheckCanceledCtx(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	got, err := (&TLS{
		Port:    443,
		Timeout: time.Second,
	}).Check(ctx, "127.0.0.1")

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Check() error = %v, want context.Canceled", err)
	}
	if got.Healthy {
		t.Error("Check() Healthy = true, want false")
	}
}

func TestTLSCheckUntrustedCert(t *testing.T) {
	server := httptest.NewUnstartedServer(http.NotFoundHandler())
	server.Config.ErrorLog = log.New(io.Discard, "", 0)
	server.StartTLS()
	defer server.Close()

	host, portText, err := net.SplitHostPort(server.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}

	got, err := (&TLS{
		Port:    port,
		Timeout: time.Second,
	}).Check(context.Background(), host)

	if err == nil {
		t.Fatal("Check() error = nil, want certificate validation error")
	}
	if got.Healthy {
		t.Error("Check() Healthy = true, want false")
	}
}
