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
) (Request, error) {
	flags := flag.NewFlagSet("config", flag.ContinueOnError)
	flags.SetOutput(errOut)
	addGlobalFlags(flags, &opts)

	if err := flags.Parse(args); err != nil {
		return Request{}, err
	}

	args = flags.Args()

	if len(args) < 1 {
		return Request{}, fmt.Errorf("expected command: init|add")
	}

	switch args[0] {
	case "init":
		flags := flag.NewFlagSet("config init", flag.ContinueOnError)
		flags.SetOutput(errOut)
		addGlobalFlags(flags, &opts)
		if err := flags.Parse(args[1:]); err != nil {
			return Request{}, err
		}
		if flags.NArg() != 0 {
			return Request{}, fmt.Errorf("init cannot accept further commands")
		}
		return Request{
			Kind:       CmdInitConfig,
			ConfigPath: opts.ConfigPath,
		}, nil
	case "add":
		return parseConfigModification(errOut, args[1:], opts, "add", CmdAddConfig)
	case "edit":
		return parseConfigModification(errOut, args[1:], opts, "edit", CmdEditConfig)
	default:
		return Request{}, fmt.Errorf("unrecognized command: %q", args[0])
	}

}

func parseConfigModification(
	errOut io.Writer,
	args []string,
	opts GlobalOption,
	name string,
	kind CommandKind,
) (Request, error) {
	flags := flag.NewFlagSet("config "+name, flag.ContinueOnError)
	flags.SetOutput(errOut)
	addGlobalFlags(flags, &opts)

	if err := flags.Parse(args); err != nil {
		return Request{}, err
	}

	if flags.NArg() != 1 {
		return Request{}, fmt.Errorf(
			"usage: homelabctl config %s <service-name>",
			name,
		)
	}

	return Request{
		Kind: kind,
		Modify: ModifyConfigRequest{
			ConfigPath:  opts.ConfigPath,
			ServiceName: flags.Arg(0),
		},
	}, nil
}
