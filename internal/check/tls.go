package check

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"
	"time"
)

type TLS struct {
	Port    int
	Timeout time.Duration
	Verbose bool
}

type TLSResults struct {
	Target    string
	Subject   string
	Issuer    string
	Names     []string
	After     string
	Expires   time.Duration
	Healthy   bool
	Latency   time.Duration
	CheckedAt time.Time
}

func (c *TLS) Check(ctx context.Context, target string) (TLSResults, error) {
	config := tls.Config{ServerName: target}
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	port := strconv.Itoa(c.Port)
	host := net.JoinHostPort(target, port)

	var d net.Dialer

	netConn, err := d.DialContext(ctx, "tcp", host)
	if err != nil {
		results := TLSResults {
			Target: target,
			Healthy: false,
			CheckedAt: time.Now(),
		}
		return results, err
	}

	start := time.Now()
	tlsConn := tls.Client(netConn, &config)
	defer tlsConn.Close()

	if err := tlsConn.HandshakeContext(ctx); err != nil {
		results := TLSResults {
			Target: target,
			Healthy: false,
			CheckedAt: time.Now(),
		}
		return results, err
	}

	state := tlsConn.ConnectionState()
	cert := state.PeerCertificates[0]
	expiresIn := time.Until(cert.NotAfter)
	now := time.Now()
	validNow := !now.Before(cert.NotBefore) && now.Before(cert.NotAfter)

	if !validNow {
		results := TLSResults {
			Target:    target,
			Subject:   cert.Subject.String(),
			Issuer:    cert.Issuer.String(),
			Names:     cert.DNSNames,
			Healthy:   false,
			CheckedAt: time.Now(),
		}

		return results, nil
	}

	results := TLSResults {
		Target:    target,
		Subject:   cert.Subject.String(),
		Issuer:    cert.Issuer.String(),
		Names:     cert.DNSNames,
		After:     cert.NotAfter.String(),
		Expires:   expiresIn,
		Healthy:   true,
		Latency:   time.Since(start),
		CheckedAt: time.Now(),
	}

	return results, nil
}
