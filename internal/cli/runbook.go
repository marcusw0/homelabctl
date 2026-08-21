package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/marcusw0/homelabctl/internal/config"
	"github.com/marcusw0/homelabctl/internal/markdown"
)

type RunbookCmd struct {
	ConfigPath  string
	RunbookPath string
	ServerName  string
	Style       string
	Width       int
}

func parseRunbook(
	errOut io.Writer,
	args []string,
	opts GlobalOption,
) (Command, error) {
	cmd := &RunbookCmd{}
	flags := flag.NewFlagSet("runbook", flag.ContinueOnError)
	flags.SetOutput(errOut)
	addGlobalFlags(flags, &opts)

	flags.IntVar(
		&cmd.Width,
		"width",
		100,
		"render width",
	)

	flags.IntVar(
		&cmd.Width,
		"w",
		100,
		"render width",
	)

	flags.StringVar(
		&cmd.Style,
		"style",
		"dark",
		"render style: ascii|dark|dracula|light|notty|pink|tokyo-night",
	)

	flags.StringVar(
		&cmd.Style,
		"s",
		"dark",
		"render style",
	)

	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	args = flags.Args()

	if len(args) != 1 {
		return nil, errors.New("runbook accepts exactly one server argument")
	}

	cmd.ConfigPath = opts.ConfigPath
	cmd.ServerName = args[0]

	return cmd, nil

}

func (c *RunbookCmd) Validate() error {
	if c.Width <= 0 {
		return errors.New("width must be a positive number")
	}

	if c.ServerName == "" {
		return errors.New("server name cannot be blank")
	}

	cfg, err := config.Load(c.ConfigPath)
	if err != nil {
		return err
	}

	server, exists := cfg.Servers[c.ServerName]
	if !exists {
		return fmt.Errorf(
			"server %q not found in %s",
			c.ServerName,
			c.ConfigPath,
		)
	}

	if server.Runbook == "" {
		return fmt.Errorf("no runbook path configured for %s", c.ServerName)
	}

	c.RunbookPath = server.Runbook

	return nil
}

func (c *RunbookCmd) Run(ctx context.Context, streams IOStreams) error {
	runbook := c.RunbookPath

	if !filepath.IsAbs(runbook) {
		runbook = filepath.Join(
			filepath.Dir(c.ConfigPath),
			runbook,
		)
	}

	file, err := os.ReadFile(runbook)
	if err != nil {
		return fmt.Errorf("read runbook: %w", err)
	}

	render, err := markdown.Render(file, c.Width, c.Style)
	if err != nil {
		return err
	}

	_, err = streams.Out.Write(render)
	if err != nil {
		return err
	}

	return nil
}
