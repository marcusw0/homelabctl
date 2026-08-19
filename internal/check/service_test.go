package check

import (
	"context"
	"io"
	"log"
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

	server := httptest.NewUnstartedServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	))
	server.Config.ErrorLog = log.New(io.Discard, "", 0)
	server.StartTLS()

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

func TestServiceCheckRunsConcurrently(t *testing.T) {
	started := make(chan string, 4)
	release := make(chan struct{})

	var releaseOnce sync.Once
	releaseChecks := func() {
		releaseOnce.Do(func() {
			close(release)
		})
	}
	defer releaseChecks()

	wait := func(name string) {
		started <- name
		<-release
	}

	checks := serviceChecks{
		HTTP: func(context.Context) (HTTPResults, error) {
			wait("HTTP")
			return HTTPResults{Healthy: true}, nil
		},
		DNS: func(context.Context) (DNSResults, error) {
			wait("DNS")
			return DNSResults{Healthy: true}, nil
		},
		TCP: func(context.Context) (TCPResults, error) {
			wait("TCP")
			return TCPResults{Healthy: true}, nil
		},
		TLS: func(context.Context) (TLSResults, error) {
			wait("TLS")
			return TLSResults{Healthy: true}, nil
		},
	}

	type response struct {
		results ServiceResults
		err     error
	}

	done := make(chan response, 1)
	go func() {
		results, err := runServiceChecks(context.Background(), checks)
		done <- response{results: results, err: err}
	}()

	seen := make(map[string]bool)
	timeout := time.NewTimer(time.Second)
	defer timeout.Stop()

	for range 4 {
		select {
		case name := <-started:
			seen[name] = true
		case <-timeout.C:
			t.Fatalf(
				"only %d checks started before blocking; checks may be sequential",
				len(seen),
			)
		}
	}

	releaseChecks()

	select {
	case got := <-done:
		if got.err != nil {
			t.Fatalf("runServiceChecks() error = %v", got.err)
		}
		if !got.results.Healthy() {
			t.Error("runServiceChecks() returned unhealthy results")
		}
	case <-time.After(time.Second):
		t.Fatal("runServiceChecks() did not return")
	}
}
