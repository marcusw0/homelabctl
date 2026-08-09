package cli

import (
	"errors"
	"fmt"
)

func parseCheck(args []string, opts GlobalOption) (Command, error) {
	if len(args) < 1 {
		return nil, errors.New("Expected subcommand: dns|http|tcp|tls|service")
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
	case "service":
		return parseCheckService(args[1:], opts)
	default:
		if len(args) != 1 {
			return nil, errors.New("check accepts exactly one server name")
		}
		return nil, fmt.Errorf("unrecognized subcommand: %q", args[0])
	}
}
