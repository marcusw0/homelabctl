package config

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/marcusw0/homelabctl/internal/check"
)

func TestServerEffectiveDefaults(t *testing.T) {
	server := Server{
		FQDN: "example.com",
		Port: 443,
	}

	wantChecks := []check.Kind{
		check.KindHTTP,
		check.KindDNS,
		check.KindTCP,
		check.KindTLS,
	}
	if got := server.EffectiveChecks(); !reflect.DeepEqual(got, wantChecks) {
		t.Errorf("EffectiveChecks() = %v, want %v", got, wantChecks)
	}
	if got := server.EffectiveHTTPURL(); got != "https://example.com:443" {
		t.Errorf("EffectiveHTTPURL() = %q, want %q", got, "https://example.com:443")
	}
	if got := server.EffectiveStatusCode(); got != http.StatusOK {
		t.Errorf("EffectiveStatusCode() = %d, want %d", got, http.StatusOK)
	}
	if got := server.EffectiveFollowRedirects(); !got {
		t.Error("EffectiveFollowRedirects() = false, want true")
	}
	if got := server.EffectiveTimeout(); got != 5*time.Second {
		t.Errorf("EffectiveTimeout() = %s, want %s", got, 5*time.Second)
	}
	if got := server.EffectiveTLSWarnBefore(); got != 15*24*time.Hour {
		t.Errorf(
			"EffectiveTLSWarnBefore() = %s, want %s",
			got,
			15*24*time.Hour,
		)
	}
}

func TestServerEffectiveOverrides(t *testing.T) {
	followRedirects := false
	tlsWarnBefore := time.Duration(0)
	server := Server{
		Checks:          []check.Kind{check.KindHTTP},
		HTTPURL:         "http://example.com/health",
		ExpectedStatus:  http.StatusNoContent,
		FollowRedirects: &followRedirects,
		Timeout:         2 * time.Second,
		TLSWarnBefore:   &tlsWarnBefore,
	}

	if got := server.EffectiveChecks(); !reflect.DeepEqual(got, server.Checks) {
		t.Errorf("EffectiveChecks() = %v, want %v", got, server.Checks)
	}
	if got := server.EffectiveHTTPURL(); got != server.HTTPURL {
		t.Errorf("EffectiveHTTPURL() = %q, want %q", got, server.HTTPURL)
	}
	if got := server.EffectiveStatusCode(); got != server.ExpectedStatus {
		t.Errorf("EffectiveStatusCode() = %d, want %d", got, server.ExpectedStatus)
	}
	if got := server.EffectiveFollowRedirects(); got {
		t.Error("EffectiveFollowRedirects() = true, want false")
	}
	if got := server.EffectiveTimeout(); got != server.Timeout {
		t.Errorf("EffectiveTimeout() = %s, want %s", got, server.Timeout)
	}
	if got := server.EffectiveTLSWarnBefore(); got != 0 {
		t.Errorf("EffectiveTLSWarnBefore() = %s, want 0s", got)
	}
}
