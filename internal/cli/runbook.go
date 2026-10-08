package cli

import (
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

func parseRunbook(
	errOut io.Writer,
	args []string,
	opts GlobalOption,
) (Request, error) {
	cmd := RunbookRequest{}
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
		return Request{}, err
	}

	args = flags.Args()

	if len(args) != 1 {
		return Request{},
			errors.New(
				"runbook accepts exactly one service argument",
			)
	}

	cmd.ConfigPath = opts.ConfigPath
	cmd.ServiceName = args[0]

	return Request{
		Kind:    CmdRunbook,
		Runbook: cmd,
	}, nil
}

func runRunbook(request RunbookRequest, streams IOStreams) error {
	if request.Width <= 0 {
		return errors.New("width must be a positive number")
	}

	if request.ServiceName == "" {
		return errors.New("service name cannot be blank")
	}

	cfg, err := config.Load(request.ConfigPath)
	if err != nil {
		return err
	}

	service, exists := cfg.Services[request.ServiceName]
	if !exists {
		return fmt.Errorf(
			"service %q not found in %s",
			request.ServiceName,
			request.ConfigPath,
		)
	}

	if service.Runbook == "" {
		return fmt.Errorf("no runbook path configured for %s", request.ServiceName)
	}

	runbook := service.Runbook

	if !filepath.IsAbs(runbook) {
		runbook = filepath.Join(
			filepath.Dir(request.ConfigPath),
			runbook,
		)
	}

	file, err := os.ReadFile(runbook)
	if err != nil {
		return fmt.Errorf("read runbook: %w", err)
	}

	render, err := markdown.Render(file, request.Width, request.Style)
	if err != nil {
		return err
	}

	return ui.Run(
		streams.In,
		streams.Out,
		runbook,
		string(render),
		request.Style,
	)
}
