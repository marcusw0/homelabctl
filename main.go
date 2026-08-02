package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/netip"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/marcusw0/homelabctl/check"
)

func httpRequest(target string) {
	parsedURL, err := url.Parse(target)
	if err != nil {
		log.Fatal(err)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		log.Fatalf("Invalid URL: %s\n", target)
	}

	ctx := context.Background()
	result, err := check.CheckHTTP(ctx, target)
	if err != nil {
		log.Fatalf(
			"FAILED\nHost: %s\nHealthy: %t\nChecked At: %v\nError: %v\n",
			parsedURL.Host,
			result.Healthy,
			result.CheckedAt.Local(),
			err,
		)
	}

	fmt.Printf(
		"Server: %s\nStatus: %d\nLatency: %dms\nHealthy: %t\nChecked At: %v\n",
		parsedURL.Host,
		result.StatusCode,
		result.Latency.Milliseconds(),
		result.Healthy,
		result.CheckedAt.Local(),
	)

	if !result.Healthy {
		fmt.Printf("Body: %s\n", result.Body)
	}

	os.Exit(0)
}

func tcpRequest(target string) {
	_, err := netip.ParseAddrPort(target)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	ctx := context.Background()
	result, err := check.CheckTCP(ctx, target)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(
		"Host: %s\nLatency: %dms\nHealthy: %t\nChecked At: %v\n",
		result.Target,
		result.Latency.Milliseconds(),
		result.Healthy,
		result.CheckedAt.Local(),
	)

	os.Exit(0)
}

func tlsRequest(target string) {
	_, _, err := net.SplitHostPort(target)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	ctx := context.Background()
	result, err := check.CheckTLS(ctx, target)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(
		"Host: %s\nSubject: %s\nIssuer: %s\nName: %s\nNotAfter: %s\nExpires: %v\nLatency: %dms\nHealthy: %t\nChecked At: %v\n",
		result.Target,
		result.Subject,
		result.Issuer,
		result.Names,
		result.After,
		formatExpiry(result.Expires),
		result.Latency.Milliseconds(),
		result.Healthy,
		result.CheckedAt.Local(),
	)

	os.Exit(0)
}

func formatExpiry(d time.Duration) string {
	days := int(d / (24 * time.Hour))
	hours := int((d % (24 * time.Hour)) / time.Hour)

	return fmt.Sprintf("%dd %dh", days, hours)
}

func dnsRequest(target string) {
	hostname := target

	if strings.Contains(target, "://") {
		parsed, err := url.Parse(target)
		if err != nil {
			// handle error
		}

		hostname = parsed.Hostname()
	}

	ctx := context.Background()
	result, err := check.CheckDNS(ctx, hostname)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(
		"Host: %s\nResponse: %v\nLatency: %dms\nHealthy: %t\nChecked At: %v\n",
		result.Target,
		result.Response,
		result.Latency.Milliseconds(),
		result.Healthy,
		result.CheckedAt.Local(),
	)

	os.Exit(0)
}

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: homelabctl check <http|tcp|tls|dns> <target>")
		os.Exit(1)
	}

	command := os.Args[1]
	protocol := os.Args[2]
	target := os.Args[3]

	if command != "check" {
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}

	switch protocol {
	case "http":
		httpRequest(target)
	case "tcp":
		tcpRequest(target)
	case "tls":
		tlsRequest(target)
	case "dns":
		dnsRequest(target)
	default:
		fmt.Printf("Unknown protocol: %s\n", protocol)
		os.Exit(1)
	}
}
