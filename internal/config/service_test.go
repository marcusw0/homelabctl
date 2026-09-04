package config

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/marcusw0/homelabctl/internal/check"
)

func TestServiceEffectiveDefaults(t *testing.T) {
	service := Service{
		FQDN: "example.com",
		Port: 443,
	}

	wantChecks := []check.Kind{
		check.KindHTTP,
		check.KindDNS,
		check.KindTCP,
		check.KindTLS,
	}
	if got := service.EffectiveChecks(); !reflect.DeepEqual(got, wantChecks) {
		t.Errorf("EffectiveChecks() = %v, want %v", got, wantChecks)
	}
	if got := service.EffectiveHTTPURL(); got != "https://example.com:443" {
		t.Errorf("EffectiveHTTPURL() = %q, want %q", got, "https://example.com:443")
	}
	if got := service.EffectiveStatusCode(); got != http.StatusOK {
		t.Errorf("EffectiveStatusCode() = %d, want %d", got, http.StatusOK)
	}
	if got := service.EffectiveFollowRedirects(); !got {
		t.Error("EffectiveFollowRedirects() = false, want true")
	}
	if got := service.EffectiveTimeout(); got != 5*time.Second {
		t.Errorf("EffectiveTimeout() = %s, want %s", got, 5*time.Second)
	}
	if got := service.EffectiveTLSWarnBefore(); got != 15*24*time.Hour {
		t.Errorf(
			"EffectiveTLSWarnBefore() = %s, want %s",
			got,
			15*24*time.Hour,
		)
	}
}

func TestServiceEffectiveOverrides(t *testing.T) {
	followRedirects := false
	tlsWarnBefore := time.Duration(0)
	service := Service{
		Checks:          []check.Kind{check.KindHTTP},
		HTTPURL:         "http://example.com/health",
		ExpectedStatus:  http.StatusNoContent,
		FollowRedirects: &followRedirects,
		Timeout:         2 * time.Second,
		TLSWarnBefore:   &tlsWarnBefore,
	}

	if got := service.EffectiveChecks(); !reflect.DeepEqual(got, service.Checks) {
		t.Errorf("EffectiveChecks() = %v, want %v", got, service.Checks)
	}
	if got := service.EffectiveHTTPURL(); got != service.HTTPURL {
		t.Errorf("EffectiveHTTPURL() = %q, want %q", got, service.HTTPURL)
	}
	if got := service.EffectiveStatusCode(); got != service.ExpectedStatus {
		t.Errorf("EffectiveStatusCode() = %d, want %d", got, service.ExpectedStatus)
	}
	if got := service.EffectiveFollowRedirects(); got {
		t.Error("EffectiveFollowRedirects() = true, want false")
	}
	if got := service.EffectiveTimeout(); got != service.Timeout {
		t.Errorf("EffectiveTimeout() = %s, want %s", got, service.Timeout)
	}
	if got := service.EffectiveTLSWarnBefore(); got != 0 {
		t.Errorf("EffectiveTLSWarnBefore() = %s, want 0s", got)
	}
}
