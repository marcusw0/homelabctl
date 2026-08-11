package cli

import (
	"flag"
	"fmt"
	"io"
)

func parseConfig(
	errOut io.Writer,
	args []string,
	opts GlobalOption,
) (Command, error) {
	flags := flag.NewFlagSet("config", flag.ContinueOnError)
	flags.SetOutput(errOut)
	addGlobalFlags(flags, &opts)

	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	args = flags.Args()

	if len(args) < 1 {
		return nil, fmt.Errorf("expected command: init|add")
	}

	switch args[0] {
	case "init":
		flags := flag.NewFlagSet("config init", flag.ContinueOnError)
		flags.SetOutput(errOut)
		addGlobalFlags(flags, &opts)
		if err := flags.Parse(args[1:]); err != nil {
			return nil, err
		}
		if flags.NArg() != 0 {
			return nil, fmt.Errorf("init cannot accept further commands")
		}
		return &ConfigInitCmd{
			ConfigPath: opts.ConfigPath,
		}, nil
	case "add":
		flags := flag.NewFlagSet("config add", flag.ContinueOnError)
		flags.SetOutput(errOut)
		addGlobalFlags(flags, &opts)
		if err := flags.Parse(args[1:]); err != nil {
			return nil, err
		}
		if flags.NArg() != 1 {
			return nil, fmt.Errorf("usage: homelabctl config add <server-name>")
		}
		return &ConfigAddCmd{
			ConfigPath: opts.ConfigPath,
			ServerName: flags.Arg(0),
		}, nil
	default:
		return nil, fmt.Errorf("unrecognized command: %q", args[0])
	}

}
