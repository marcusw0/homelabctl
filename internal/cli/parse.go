package cli

import (
	"flag"
	"fmt"
	"io"

	"github.com/marcusw0/homelabctl/internal/config"
)

func Parse(args []string, errOut io.Writer) (Request, error) {

	cfgDir, err := config.DefaultPath()
	if err != nil {
		return Request{}, err
	}
	opts := GlobalOption{ConfigPath: cfgDir}

	flags := flag.NewFlagSet("homelabctl", flag.ContinueOnError)
	flags.SetOutput(errOut)
	addGlobalFlags(flags, &opts)

	flags.Usage = func() {
		fmt.Fprintln(errOut, `Usage:
   homelabctl [global options] <command> [command options]

Commands:
   check     Run HTTP, TCP, TLS, DNS, or service check from config
   config    Initialize or modify service config
   list      List services in your config.toml
   runbook   View/edit service runbooks
   help      Show this help

Example:
   homelabctl check http myservice.example.com
   homelabctl -v check service myservice
   homelabctl config init

Global options:`)

		flags.PrintDefaults()
		fmt.Fprintln(errOut, "\nFor help with a specific command: homelabctl <command> --help")
	}

	if err := flags.Parse(args); err != nil {
		return Request{}, err
	}
	args = flags.Args()
	if len(args) == 0 || args[0] == "help" {
		flags.Usage()
		return Request{}, flag.ErrHelp
	}

	return parseCommand(args, opts, errOut)
}

func parseCommand(
	args []string,
	opts GlobalOption,
	errOut io.Writer,
) (Request, error) {
	switch args[0] {
	case "check":
		return parseCheck(errOut, args[1:], opts)
	case "config":
		return parseConfig(errOut, args[1:], opts)
	case "list":
		return parseList(errOut, args[1:], opts)
	case "runbook":
		return parseRunbook(errOut, args[1:], opts)
	case "dashboard":
		return Request{
			Kind:       CmdDashboard,
			ConfigPath: opts.ConfigPath,
		}, nil
	default:
		return Request{}, fmt.Errorf("Unknown command: %q", args[0])
	}
}
