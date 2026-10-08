package cli

import (
	"errors"
	"fmt"

	"github.com/marcusw0/homelabctl/internal/config"
)

func runConfigInit(path string, streams IOStreams) error {
	if path == "" {
		return errors.New("config path cannot be empty")
	}
	if err := config.Initialize(path); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		streams.Out,
		"config initialized at %s\n",
		path,
	); err != nil {
		return err
	}
	return nil
}
