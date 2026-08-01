package check

import (
	"context"
	"crypto/tls"
	"net"
	"time"
)

type CertDetails struct {
	Subject   string
	Issuer    string
	Name      string
	After     string
	Target    string
	Healthy   bool
	Latency   time.Duration
	CheckedAt time.Time
}

func CheckTLS(ctx context.Context, target string) (CertDetails, error) {
	host, _, err := net.SplitHostPort(target)
	if err != nil {
		return CertDetails{}, err
	}

	config := tls.Config{ServerName: host}

	ctx, cancel := context.WithTimeout(ctx, 150*time.Millisecond)
	defer cancel()

	var d net.Dialer

	netConn, err := d.DialContext(ctx, "tcp", target)
	if err != nil {
		return CertDetails{}, err
	}

	start := time.Now()
	tlsConn := tls.Client(netConn, &config)
	defer tlsConn.Close()

	if err := tlsConn.HandshakeContext(ctx); err != nil {
		return CertDetails{}, err
	}

	state := tlsConn.ConnectionState()
	cert := state.PeerCertificates[0]
	result := CertDetails{
		Subject:   cert.Subject.String(),
		Issuer:    cert.Issuer.String(),
		Name:      cert.DNSNames[0],
		After:     cert.NotAfter.String(),
		Target:    target,
		Healthy:   true,
		Latency:   time.Since(start),
		CheckedAt: time.Now(),
	}

	return result, nil
}
