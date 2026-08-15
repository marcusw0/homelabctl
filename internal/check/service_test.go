package check

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func newTestService(t *testing.T) (*Service, func()) {
	t.Helper()

	server := httptest.NewTLSServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	))

	_, portText, err := net.SplitHostPort(server.Listener.Addr().String())
	if err != nil {
		server.Close()
		t.Fatal(err)
	}

	port, err := strconv.Atoi(portText)
	if err != nil {
		server.Close()
		t.Fatal(err)
	}

	service := &Service{
		FQDN:    "localhost",
		IP:      "127.0.0.1",
		Port:    port,
		Timeout: time.Second,
	}

	return service, server.Close
}

func TestServiceCheck(t *testing.T) {
	service, closeServer := newTestService(t)
	defer closeServer()

	got, err := service.Check(context.Background())

	if !got.DNS.Healthy {
		t.Error("DNS result unhealthy, want healthy")
	}
	if !got.TCP.Healthy {
		t.Error("TCP result unhealthy, want healthy")
	}
	if got.HTTP.Healthy {
		t.Error("HTTP result healthy, want unhealthy")
	}
	if got.TLS.Healthy {
		t.Error("TLS result healthy, want unhealthy")
	}

	if err == nil {
		t.Fatal("Check() error = nil, want HTTP and TLS errors")
	}
	if !strings.Contains(err.Error(), "HTTP check:") {
		t.Errorf("Check() error does not contain HTTP label: %v", err)
	}
	if !strings.Contains(err.Error(), "TLS check:") {
		t.Errorf("Check() error does not contain TLS label: %v", err)
	}
}

func TestServiceCheckConcurrentCalls(t *testing.T) {
	service, closeServer := newTestService(t)
	defer closeServer()

	const calls = 10

	start := make(chan struct{})
	failures := make(chan string, calls)
	var wg sync.WaitGroup

	wg.Add(calls)
	for range calls {
		go func() {
			defer wg.Done()
			<-start

			got, err := service.Check(context.Background())
			if err == nil {
				failures <- "Check() error = nil"
				return
			}
			if !got.DNS.Healthy || !got.TCP.Healthy {
				failures <- "DNS or TCP result was unhealthy"
			}
		}()
	}
	close(start)
	wg.Wait()
	close(failures)

	for failure := range failures {
		t.Error(failure)
	}
}
