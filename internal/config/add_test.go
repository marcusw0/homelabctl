package config

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/marcusw0/homelabctl/internal/check"
)

func newTestConfig(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.toml")

	if err := Initialize(path); err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}

	return path
}

func validTestServer() Server {
	return Server{
		FQDN:    "myserver.example.com",
		IP:      "192.168.5.13",
		Port:    443,
		Enabled: true,
		Runbook: "runbooks/myserver.md",
	}
}

func TestAddServer(t *testing.T) {
	cfgPath := newTestConfig(t)
	want := validTestServer()

	if err := AddServer(cfgPath, "myserver", want); err != nil {
		t.Fatalf("AddServer() error: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	got, exists := cfg.Servers["myserver"]
	if !exists {
		t.Fatal(`server "myserver" was not added`)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("server: %#v, want: %#v", got, want)
	}
}

func TestAddServerOptionalFieldsRoundTrip(t *testing.T) {
	cfgPath := newTestConfig(t)
	followRedirects := false
	tlsWarnBefore := 30 * 24 * time.Hour
	want := validTestServer()
	want.Checks = []check.Kind{check.KindHTTP, check.KindTLS}
	want.HTTPURL = "https://myserver.example.com/health"
	want.ExpectedStatus = 204
	want.FollowRedirects = &followRedirects
	want.TLSWarnBefore = &tlsWarnBefore
	want.Timeout = 3 * time.Second

	if err := AddServer(cfgPath, "myserver", want); err != nil {
		t.Fatalf("AddServer() error: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if got := cfg.Servers["myserver"]; !reflect.DeepEqual(got, want) {
		t.Errorf("server: %#v, want: %#v", got, want)
	}
}

func TestAddRejectsDup(t *testing.T) {
	cfgPath := newTestConfig(t)
	server := validTestServer()

	if err := AddServer(cfgPath, "myserver", server); err != nil {
		t.Fatalf("AddServer() error: %v", err)
	}

	before, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	err = AddServer(cfgPath, "myserver", server)
	if err == nil {
		t.Fatal("Duplicate server was added with AddServer(), want error")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("unexpected error: %v", err)
	}

	after, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(before, after) {
		t.Error("config changed after duplicate was rejected")
	}
}

func TestAddRejectsInvalidNoConfigChange(t *testing.T) {
	negativeDuration := -time.Second

	tests := []struct {
		name   string
		server Server
	}{
		{
			name: "invalid hostname",
			server: Server{
				FQDN: "bad_host",
				IP:   "192.168.5.13",
				Port: 443,
			},
		},
		{
			name: "invalid ip",
			server: Server{
				FQDN: "myserver.example.com",
				IP:   "not-an-ip",
				Port: 443,
			},
		},
		{
			name: "invalid port",
			server: Server{
				FQDN: "myserver.example.com",
				IP:   "192.168.5.13",
				Port: 0,
			},
		},
		{
			name: "negative timeout",
			server: Server{
				FQDN:    "myserver.example.com",
				IP:      "192.168.5.13",
				Port:    443,
				Timeout: negativeDuration,
			},
		},
		{
			name: "negative TLS warning duration",
			server: Server{
				FQDN:          "myserver.example.com",
				IP:            "192.168.5.13",
				Port:          443,
				TLSWarnBefore: &negativeDuration,
			},
		},
		{
			name: "unsupported check",
			server: Server{
				FQDN:   "myserver.example.com",
				IP:     "192.168.5.13",
				Port:   443,
				Checks: []check.Kind{"htpt"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgPath := newTestConfig(t)

			before, err := os.ReadFile(cfgPath)
			if err != nil {
				t.Fatal(err)
			}

			if err := AddServer(cfgPath, "myserver", tt.server); err == nil {
				t.Fatal("AddServer() error = nil, want validation error")
			}

			after, err := os.ReadFile(cfgPath)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(before, after) {
				t.Error("config changed after invalid server was rejected")
			}
		})
	}
}
