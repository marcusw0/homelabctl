package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/marcusw0/homelabctl/internal/config"
)

type Command interface {
	Validate() error
	Run(context.Context) error
}

type GlobalOption struct {
	ConfigPath string
	Verbose    bool
}

func Parse(args []string) (Command, error) {

	opts := GlobalOption{}

	cfgDir, err := config.DefaultPath()
	if err != nil {
		return nil, err
	}

	flags := flag.NewFlagSet("homelabctl", flag.ContinueOnError)
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
		return nil, errors.New("Expected a command: check|config|list")
	}

	switch args[0] {
	case "check":
		return parseCheck(args[1:])
	case "config":
		return parseConfig(args[1:], opts)
	// case "list":
	// 	return parseList(args[1:])
	default:
		return nil, fmt.Errorf("Unknown command: %q", args[0])
	}
}
