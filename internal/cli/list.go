package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	"github.com/marcusw0/homelabctl/internal/config"
)

type ListCmd struct {
	ConfigPath string
}

func parseList(
	errOut io.Writer,
	args []string,
	opts GlobalOption,
) (Command, error) {
	flags := flag.NewFlagSet("list", flag.ContinueOnError)
	flags.SetOutput(errOut)
	addGlobalFlags(flags, &opts)

	if err := flags.Parse(args); err != nil {
		return nil, err
	}
	if flags.NArg() != 0 {
		return nil, errors.New("list does not accept further arguments")
	}

	return &ListCmd{
		ConfigPath: opts.ConfigPath,
	}, nil
}

func (c *ListCmd) Validate() error {
	if c.ConfigPath == "" {
		return errors.New("config path cannot be empty")
	}
	return nil
}

func (c *ListCmd) Run(ctx context.Context, streams IOStreams) error {
	cfg, err := config.Load(c.ConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if err := writeListResponse(streams.Out, cfg); err != nil {
		return err
	}

	return nil
}

func writeListResponse(out io.Writer, resp config.Config) error {
	if len(resp.Servers) == 0 {
		_, err := fmt.Fprintln(out, "No services configured.")
		return err
	}

	writer := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)

	if _, err := fmt.Fprintln(
		writer,
		"NAME\tFQDN\tIP\tPORT\tENABLED",
	); err != nil {
		return err
	}

	names := make([]string, 0, len(resp.Servers))
	for name := range resp.Servers {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		server := resp.Servers[name]
		if _, err := fmt.Fprintf(
			writer,
			"%s\t%s\t%s\t%d\t%t\n",
			name,
			server.FQDN,
			server.IP,
			server.Port,
			server.Enabled,
		); err != nil {
			return err
		}
	}

	return writer.Flush()
}
