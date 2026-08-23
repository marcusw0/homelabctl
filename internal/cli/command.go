package cli

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/marcusw0/homelabctl/internal/config"
)

type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}

type Command interface {
	Validate() error
	Run(context.Context, IOStreams) error
}

type GlobalOption struct {
	ConfigPath string
	Verbose    bool
}

func addGlobalFlags(flags *flag.FlagSet, opts *GlobalOption) {
	flags.StringVar(
		&opts.ConfigPath,
		"config",
		opts.ConfigPath,
		"path to config file",
	)

	flags.BoolVar(
		&opts.Verbose,
		"v",
		opts.Verbose,
		"verbose output",
	)

	flags.BoolVar(
		&opts.Verbose,
		"verbose",
		opts.Verbose,
		"verbose output",
	)
}

func Parse(args []string, errOut io.Writer) (Command, error) {

	cfgDir, err := config.DefaultPath()
	if err != nil {
		return nil, err
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
   homelabctl check http myserver.example.com
   homelabctl -v check service myserver
   homelabctl config init

Global options:`)

		flags.PrintDefaults()
		fmt.Fprintln(errOut, "\nFor help with a specific command: homelabctl <command> --help")
	}

	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	args = flags.Args()

	if len(args) == 0 {
		flags.Usage()
		return nil, flag.ErrHelp
	}

	switch args[0] {
	case "help":
		flags.Usage()
		return nil, flag.ErrHelp
	case "check":
		return parseCheck(errOut, args[1:], opts)
	case "config":
		return parseConfig(errOut, args[1:], opts)
	case "list":
		return parseList(errOut, args[1:], opts)
	case "runbook":
		return parseRunbook(errOut, args[1:], opts)
	default:
		return nil, fmt.Errorf("Unknown command: %q", args[0])
	}
}
