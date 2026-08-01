package check

import (
	"context"
	"crypto/tls"
	"net"
	"time"
)

type CertDetails struct {
	Subject string
	Issuer string
	Name string
	Before string
	After string
}

func CheckTLS (ctx context.Context, target string) (CertDetails, error) {
	host, _, err := net.SplitHostPort(target)
	if err != nil {
		return CertDetails{}, err
	}

	config := tls.Config{ ServerName: host}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var d net.Dialer

	netConn, err := d.DialContext(ctx, "tcp", target)
	if err != nil {
		return CertDetails{}, err
	}

	tlsConn := tls.Client(netConn, &config)
	defer tlsConn.Close()

	if err := tlsConn.HandshakeContext(ctx); err != nil {
		return CertDetails{}, err
	}

	state := tlsConn.ConnectionState()
	cert := state.PeerCertificates[0]
	result := CertDetails {
		cert.Subject.String(),
		cert.Issuer.String(),
		cert.DNSNames[0],
		cert.NotBefore.String(),
		cert.NotAfter.String(),
	}

	return result, nil
}
