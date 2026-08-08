package cli

import "fmt"

type ConfigInitCommand struct {
	ConfigPath string
}

func parseConfig(args []string, opts GlobalOption) (Command, error) {

	if len(args) <= 0 {
		return nil, fmt.Errorf("expected command: init|list|add")
	}

	switch args[0] {
	case "init":
		return &ConfigInitCommand{
			ConfigPath: opts.ConfigPath,
		}, nil

	default:
		return nil, fmt.Errorf("unrecognized command: %q", args[0])
	}

}
