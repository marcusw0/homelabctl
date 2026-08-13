package cli

import (
	"context"
	"errors"
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

	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	args = flags.Args()

	if len(args) == 0 {
		return nil, errors.New("Expected a command: check|config|list")
	}

	switch args[0] {
	case "check":
		return parseCheck(errOut, args[1:], opts)
	case "config":
		return parseConfig(errOut, args[1:], opts)
	case "list":
		return parseList(errOut, args[1:], opts)
	default:
		return nil, fmt.Errorf("Unknown command: %q", args[0])
	}
}
