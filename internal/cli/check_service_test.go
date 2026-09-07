package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/marcusw0/homelabctl/internal/check"
	"github.com/marcusw0/homelabctl/internal/config"
)

type testTransport func(*http.Request) (*http.Response, error)

func (f testTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// Inspect the actual request deadline without sleeps or network timing assumptions.
func TestConfiguredHTTPCheck(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(`[services.web]
fqdn = "localhost"
ip = "127.0.0.1"
port = 1
enabled = true
checks = ["http"]
http_url = "http://example.com/health"
expect_status = 404
timeout = "3s"
`), 0o600); err != nil {
		t.Fatal(err)
	}
	original := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = original })
	for _, tt := range []struct {
		name    string
		args    []string
		timeout time.Duration
	}{
		{"configured", nil, 3 * time.Second},
		{"override", []string{"--timeout", "7s"}, 7 * time.Second},
	} {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			http.DefaultTransport = testTransport(func(r *http.Request) (*http.Response, error) {
				called = true
				deadline, ok := r.Context().Deadline()
				remaining := time.Until(deadline)
				if !ok || remaining > tt.timeout || remaining < tt.timeout-time.Second {
					t.Errorf("request timeout = %s, want about %s", remaining, tt.timeout)
				}
				if r.URL.String() != "http://example.com/health" {
					t.Errorf("URL = %s", r.URL)
				}
				return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
			})
			cmd, err := parseServiceCheck(io.Discard, append(tt.args, "web"), GlobalOption{ConfigPath: path})
			if err != nil {
				t.Fatal(err)
			}
			if err := cmd.Validate(); err != nil {
				t.Fatal(err)
			}
			var out bytes.Buffer
			if err := cmd.Run(context.Background(), IOStreams{Out: &out, ErrOut: io.Discard}); err != nil {
				t.Fatal(err)
			}
			if !called || !strings.Contains(out.String(), "HTTP status: healthy") || strings.Count(out.String(), "status: skipped") != 3 {
				t.Fatalf("unexpected selected-check output: %s", &out)
			}
		})
	}
	for _, timeout := range []string{"0s", "-1s"} {
		cmd, err := parseServiceCheck(io.Discard, []string{"--timeout", timeout, "web"}, GlobalOption{ConfigPath: path})
		if err != nil {
			t.Fatal(err)
		}
		if err := cmd.Validate(); err == nil {
			t.Errorf("accepted timeout %s", timeout)
		}
	}
}

func TestServiceFailureResults(t *testing.T) {
	original := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = original })
	transportErr := errors.New("connection failed")
	service := config.Service{FQDN: "localhost", IP: "127.0.0.1", Port: 1, Checks: []check.Kind{check.KindHTTP}, HTTPURL: "http://example.com"}
	for _, tt := range []struct {
		name    string
		status  int
		err     error
		wantErr bool
	}{
		{"healthy", 200, nil, false},
		{"unhealthy without transport error", 503, nil, true},
		{"transport error", 0, transportErr, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			http.DefaultTransport = testTransport(func(*http.Request) (*http.Response, error) {
				if tt.err != nil {
					return nil, tt.err
				}
				return &http.Response{StatusCode: tt.status, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
			})
			for _, cmd := range []Command{&ServiceCheckCmd{serviceName: "web", serviceCfg: service}, &AllCmd{Services: map[string]config.Service{"web": service}}} {
				var out bytes.Buffer
				err := cmd.Run(context.Background(), IOStreams{Out: &out, ErrOut: io.Discard})
				if (err != nil) != tt.wantErr {
					t.Fatalf("%T error = %v", cmd, err)
				}
				if tt.err != nil && !errors.Is(err, tt.err) {
					t.Errorf("lost transport error: %v", err)
				}
				if out.Len() == 0 {
					t.Error("missing results on failure")
				}
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := (&AllCmd{}).Run(ctx, IOStreams{Out: io.Discard, ErrOut: io.Discard}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled run error = %v", err)
	}
}

func TestVerboseSkippedAndWarning(t *testing.T) {
	resp := check.ServiceResults{
		HTTP: check.Outcome[check.HTTPResults]{Status: check.StatusSkipped},
		DNS:  check.Outcome[check.DNSResults]{Status: check.StatusSkipped},
		TCP:  check.Outcome[check.TCPResults]{Status: check.StatusSkipped},
		TLS:  check.Outcome[check.TLSResults]{Status: check.StatusWarning, Result: check.TLSResults{Subject: "example.com", Expires: 24 * time.Hour}},
	}
	var out bytes.Buffer
	if err := writeService(&out, resp, "web", true); err != nil {
		t.Fatal(err)
	}
	if strings.Count(out.String(), "Health: skipped") != 3 || !strings.Contains(out.String(), "Health: warning") || !strings.Contains(out.String(), "Expires: 1d 0h") {
		t.Fatalf("unexpected output: %s", &out)
	}
	for _, phantom := range []string{"Status: 0", "Latency:", "Response:", "Message:"} {
		if strings.Contains(out.String(), phantom) {
			t.Errorf("skipped details contain %q", phantom)
		}
	}
	if !resp.Healthy() {
		t.Error("warning/skipped results should remain healthy")
	}
}
