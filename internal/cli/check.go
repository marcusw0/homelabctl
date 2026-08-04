package cli

import (
	"errors"
	"fmt"
)

func parseCheck(args []string) (Command, error) {
	if len(args) < 1 {
		return nil, errors.New("Expected command: dns|http|tcp|tls")
	}

	switch args[0] {
	case "dns":
		return parseDNSCheck(args[1:])
	case "http":
		return parseHTTPCheck(args[1:])
	case "tcp":
		return parseTCPCheck(args[1:])
	case "tls":
		return parseTLSCheck(args[1:])
	default:
		return nil, fmt.Errorf("Unknown check type: %q", args[0])
	}
}
