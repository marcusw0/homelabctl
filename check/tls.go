package check

import (
	"context"
	"crypto/tls"
	"net"
	"time"
)

type CertDetails struct {
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

func CheckTLS(ctx context.Context, target string) (CertDetails, error) {
	host, _, err := net.SplitHostPort(target)
	if err != nil {
		return CertDetails{}, err
	}

	config := tls.Config{ServerName: host}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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
	expiresIn := time.Until(cert.NotAfter)
	now := time.Now()
	validNow := !now.Before(cert.NotBefore) && now.Before(cert.NotAfter)

	if !validNow {
		result := CertDetails{
			Target:    target,
			Subject:   cert.Subject.String(),
			Issuer:    cert.Issuer.String(),
			Names:     cert.DNSNames,
			Healthy:   false,
			CheckedAt: time.Now(),
		}

		return result, nil
	}

	result := CertDetails{
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

	return result, nil
}
