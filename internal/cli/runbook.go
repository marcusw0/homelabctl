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
	ui "github.com/marcusw0/homelabctl/internal/tui/runbook"
)

type RunbookCmd struct {
	ConfigPath  string
	RunbookPath string
	ServiceName string
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
		150,
		"render width",
	)

	flags.IntVar(
		&cmd.Width,
		"w",
		150,
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
		return nil, errors.New("runbook accepts exactly one service argument")
	}

	cmd.ConfigPath = opts.ConfigPath
	cmd.ServiceName = args[0]

	return cmd, nil

}

func (c *RunbookCmd) Validate() error {
	if c.Width <= 0 {
		return errors.New("width must be a positive number")
	}

	if c.ServiceName == "" {
		return errors.New("service name cannot be blank")
	}

	cfg, err := config.Load(c.ConfigPath)
	if err != nil {
		return err
	}

	service, exists := cfg.Services[c.ServiceName]
	if !exists {
		return fmt.Errorf(
			"service %q not found in %s",
			c.ServiceName,
			c.ConfigPath,
		)
	}

	if service.Runbook == "" {
		return fmt.Errorf("no runbook path configured for %s", c.ServiceName)
	}

	c.RunbookPath = service.Runbook

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

	return ui.Run(
		streams.In,
		streams.Out,
		runbook,
		string(render),
		c.Style,
	)
}
