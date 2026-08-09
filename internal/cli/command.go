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

func Parse(args []string, errOut io.Writer) (Command, error) {

	opts := GlobalOption{}

	cfgDir, err := config.DefaultPath()
	if err != nil {
		return nil, err
	}

	flags := flag.NewFlagSet("homelabctl", flag.ContinueOnError)
	flags.SetOutput(errOut)

	flags.StringVar(
		&opts.ConfigPath,
		"config",
		cfgDir,
		"path to config file",
	)

	flags.BoolVar(
		&opts.Verbose,
		"v",
		false,
		"verbose output",
	)

	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	args = flags.Args()

	if len(args) == 0 {
		return nil, errors.New("Expected a command: check|config|list\n")
	}

	switch args[0] {
	case "check":
		return parseCheck(args[1:], opts)
	case "config":
		return parseConfig(args[1:], opts)
	// case "list":
	// 	return parseList(args[1:])
	default:
		return nil, fmt.Errorf("Unknown command: %q", args[0])
	}
}
