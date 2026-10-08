package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"

	"github.com/marcusw0/homelabctl/internal/check"
)

func parseCheck(
	errOut io.Writer,
	args []string,
	opts GlobalOption,
) (Request, error) {
	var checkAll bool

	if len(args) < 1 {
		return Request{}, errors.New("Expected subcommand: dns|http|tcp|tls|service")
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
		return Request{}, err
	}

	if checkAll == true {
		return Request{
			Kind:       CmdAllCheck,
			ConfigPath: opts.ConfigPath,
		}, nil
	}

	args = flags.Args()
	if len(args) < 1 {
		return Request{}, errors.New("Expected subcommand: dns|http|tcp|tls|service")
	}

	switch args[0] {
	case "dns":
		return parseTargetCheck(errOut, args[1:], opts, "dns", CmdDNSCheck)
	case "http":
		return parseHTTPCheck(errOut, args[1:], opts)
	case "tcp":
		return parseTargetCheck(errOut, args[1:], opts, "tcp", CmdTCPCheck)
	case "tls":
		return parseTargetCheck(errOut, args[1:], opts, "tls", CmdTLSCheck)
	case "service":
		return parseTargetCheck(errOut, args[1:], opts, "service", CmdServiceCheck)
	default:
		if len(args) != 1 {
			return Request{}, errors.New("check accepts exactly one service name")
		}
		return Request{}, fmt.Errorf("unrecognized subcommand: %q", args[0])
	}
}

func parseTargetCheck(
	errOut io.Writer,
	args []string,
	opts GlobalOption,
	name string,
	kind CommandKind,
) (Request, error) {
	flags := flag.NewFlagSet("check "+name, flag.ContinueOnError)
	flags.SetOutput(errOut)
	addGlobalFlags(flags, &opts)

	if err := flags.Parse(args); err != nil {
		return Request{}, err
	}

	if flags.NArg() != 1 {
		return Request{}, fmt.Errorf(
			"check %s accepts exactly one argument",
			name,
		)
	}

	request := Request{
		Kind: kind,
		Check: CheckRequest{
			Target:  flags.Arg(0),
			Timeout: opts.Timeout,
			Verbose: opts.Verbose,
		},
	}

	if kind == CmdServiceCheck {
		request.ConfigPath = opts.ConfigPath
	}

	return request, nil
}

func parseHTTPCheck(
	errOut io.Writer,
	args []string,
	opts GlobalOption,
) (Request, error) {
	cmd := HTTPCheckRequest{}

	flags := flag.NewFlagSet("check http", flag.ContinueOnError)
	flags.SetOutput(errOut)
	addGlobalFlags(flags, &opts)

	flags.IntVar(
		&cmd.ExpectedStatus,
		"expect-status",
		http.StatusOK,
		"Expected HTTP Status Code",
	)

	flags.BoolVar(
		&cmd.FollowRedirect,
		"follow-redirects",
		true,
		"Should redirects be followed",
	)

	if err := flags.Parse(args); err != nil {
		return Request{}, err
	}

	if flags.NArg() != 1 {
		return Request{},
			errors.New(
				"check http accepts only one argument",
			)
	}

	cmd.Target = flags.Arg(0)
	cmd.Timeout = opts.Timeout
	cmd.Verbose = opts.Verbose

	return Request{
		Kind: CmdHTTPCheck,
		HTTP: cmd,
	}, nil
}

func parseTLSTarget(target string) (string, int, error) {
	host := target
	port := 443

	if _, err := netip.ParseAddr(target); err != nil &&
		strings.Contains(target, ":") {
		var portText string
		host, portText, err = net.SplitHostPort(target)
		if err != nil {
			return "", 0, err
		}

		number, err := strconv.ParseUint(portText, 10, 16)
		if err != nil || number == 0 {
			return "", 0, errors.New("port must be between 1 and 65535")
		}
		port = int(number)
	}

	if err := check.ValidateIPOrHostname(host); err != nil {
		return "", 0, err
	}

	return host, port, nil
}
