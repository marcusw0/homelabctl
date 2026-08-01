package main

import (
	"context"
	"fmt"
	"log"
	"net/netip"
	"net/url"
	"os"
	"strings"

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
		log.Fatal(err)
	}

	fmt.Printf(
		"Server: %s\nStatus: %d\nLatency: %dms\nHealthy: %t\n",
		parsedURL.Host,
		result.StatusCode,
		result.Latency.Milliseconds(),
		result.Healthy,
	)

	if !result.Healthy {
		fmt.Printf("Body: %s\n", result.Body)
	}

	os.Exit(0)
}

func tcpRequest(target string) {
	host, err := netip.ParseAddrPort(target)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	fmt.Println(host, target)
	result, err := check.CheckTCP(target)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("TCP result: %s\n", result)
	os.Exit(0)
}

func tlsRequest(target string) {
	host, port, ok := strings.Cut(target, ":")
	if !ok {
		log.Fatalln("Missing port seperator ':'")
	}

	if port != "443" && port != "80" {
		log.Fatalf("Invalid port: %s\n", port)
	}

	if strings.HasPrefix(host, "http") {
		log.Fatalln("Do not include http prefix")
	}

	ctx := context.Background()
	result, err := check.CheckTLS(ctx,target)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(
		"Subject: %s\nIssuer: %s\nName: %s\nNotBefore: %s\nNotAfter: %s\n",
		result.Subject,
		result.Issuer,
		result.Name,
		result.Before,
		result.After,
	)

	os.Exit(0)
}

func main() {
	if len(os.Args) < 3 || len(os.Args) > 4 {
		fmt.Println("Usage: check http <url>")
		fmt.Println("Usage: check tcp <ip:port>")
		os.Exit(1)
	}

	command := os.Args[1]
	protocol := os.Args[2]
	target := os.Args[3]

	if command != "check" {
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}

	if protocol == "http" {
		httpRequest(target)
	}

	if protocol == "tcp" {
		tcpRequest(target)
	}

	if protocol == "tls" {
		tlsRequest(target)
	}

	fmt.Printf("Unknown protocol: %s\n", protocol)
	os.Exit(1)
}
