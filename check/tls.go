package check

import (
	"crypto/tls"
	"net"
)

type CertDetails struct {
	Subject string
	Issuer string
	Name string
	Before string
	After string
}

func CheckTLS (target string) (CertDetails, error) {
	host, _, err := net.SplitHostPort(target)
	if err != nil {
		return CertDetails{}, err
	}
	config := tls.Config{ ServerName: host}
	conn, err := tls.Dial("tcp", target, &config)
	if err != nil {
		return CertDetails{}, err
	}

	defer conn.Close()
	state := conn.ConnectionState()
	cert := state.PeerCertificates[0]
	result := CertDetails {
		cert.Subject.String(),
		cert.Issuer.String(),
		cert.DNSNames[0],
		cert.NotBefore.String(),
		cert.NotAfter.String(),
	}
	
	// // Read response from the server
	// buf := make([]byte, 1024)
	// n, err := conn.Read(buf)
	// if err != nil {
	// 	return "Failed to read response: ", err
	// }
	// results := string(buf[:n])
	return result, nil
}
