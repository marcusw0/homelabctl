package cli

import (
	"context"
	"errors"
	"fmt"
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
	if len(args) == 0 {
		return nil, errors.New("Expected a command: check or list")
	}

	switch args[0] {
	case "check":
		return parseCheck(args[1:])
	// case "list":
	// 	return parseList(args[1:])
	default:
		return nil, fmt.Errorf("Unknown command: %q", args[0])
	}
}
