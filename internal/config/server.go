package config

import (
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/marcusw0/homelabctl/internal/check"
)

func (s Server) EffectiveChecks() []check.Kind {
	if len(s.Checks) > 0 {
		return s.Checks
	}

	return []check.Kind{
		check.KindHTTP,
		check.KindDNS,
		check.KindTCP,
		check.KindTLS,
	}
}

func (s Server) EffectiveTCPHost() string {
	if s.IP != "" {
		return net.JoinHostPort(s.IP, strconv.Itoa(s.Port))
	}

	return net.JoinHostPort(s.FQDN, strconv.Itoa(s.Port))
}

func (s Server) EffectiveHTTPURL() string {
	if s.HTTPURL != "" {
		return s.HTTPURL
	}

	return "https://" + net.JoinHostPort(
		s.FQDN,
		strconv.Itoa(s.Port),
	)
}

func (s Server) EffectiveStatusCode() int {
	if s.ExpectedStatus == 0 {
		return http.StatusOK
	}

	return s.ExpectedStatus
}

func (s Server) EffectiveFollowRedirects() bool {
	if s.FollowRedirects == nil {
		return true
	}

	return *s.FollowRedirects
}

func (s Server) EffectiveTimeout() time.Duration {
	if s.Timeout == 0 {
		return 5 * time.Second
	}

	return s.Timeout
}

func (s Server) EffectiveTLSWarnBefore() time.Duration {
	if s.TLSWarnBefore == nil {
		return 15 * 24 * time.Hour
	}

	return *s.TLSWarnBefore
}
