package cli

import "fmt"

func parseConfig(args []string, opts GlobalOption) (Command, error) {

	if len(args) < 1 {
		return nil, fmt.Errorf("expected command: init|add")
	}

	switch args[0] {
	case "init":
		if len(args) != 1 {
			return nil, fmt.Errorf("init cannot accept further commands")
		}
		return &ConfigInitCmd{
			ConfigPath: opts.ConfigPath,
		}, nil
	case "add":
		if len(args) != 2 {
			return nil, fmt.Errorf("usage: homelabctl config add <server-name>")
		}
		return &ConfigAddCmd{
			ConfigPath: opts.ConfigPath,
			ServerName: args[1],
		}, nil
	default:
		return nil, fmt.Errorf("unrecognized command: %q", args[0])
	}

}
