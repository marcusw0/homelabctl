package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
)

func parseCheck(
	errOut io.Writer,
	args []string,
	opts GlobalOption,
) (Command, error) {
	var checkAll bool

	if len(args) < 1 {
		return nil, errors.New("Expected subcommand: dns|http|tcp|tls|service")
	}

	flags := flag.NewFlagSet("check", flag.ContinueOnError)
	flags.SetOutput(errOut)
	addGlobalFlags(flags, &opts)

	flags.BoolVar(
		&checkAll,
		"all",
		false,
		"check all services",
	)

	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	if checkAll == true {
		return &AllCmd{
			ConfigPath: opts.ConfigPath,
		}, nil
	}

	args = flags.Args()
	if len(args) < 1 {
		return nil, errors.New("Expected subcommand: dns|http|tcp|tls|service")
	}

	switch args[0] {
	case "dns":
		return parseDNSCheck(errOut, args[1:], opts)
	case "http":
		return parseHTTPCheck(errOut, args[1:], opts)
	case "tcp":
		return parseTCPCheck(errOut, args[1:], opts)
	case "tls":
		return parseTLSCheck(errOut, args[1:], opts)
	case "service":
		return parseServiceCheck(errOut, args[1:], opts)
	default:
		if len(args) != 1 {
			return nil, errors.New("check accepts exactly one server name")
		}
		return nil, fmt.Errorf("unrecognized subcommand: %q", args[0])
	}
}
