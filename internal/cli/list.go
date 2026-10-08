package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	"github.com/marcusw0/homelabctl/internal/config"
)

func parseList(
	errOut io.Writer,
	args []string,
	opts GlobalOption,
) (Request, error) {
	flags := flag.NewFlagSet("list", flag.ContinueOnError)
	flags.SetOutput(errOut)
	addGlobalFlags(flags, &opts)

	if err := flags.Parse(args); err != nil {
		return Request{}, err
	}
	if flags.NArg() != 0 {
		return Request{}, errors.New("list does not accept further arguments")
	}
	return Request{
		Kind:       CmdListConfig,
		ConfigPath: opts.ConfigPath,
	}, nil
}

func runList(path string, streams IOStreams) error {
	if path == "" {
		return errors.New("config path cannot be empty")
	}
	cfg, err := config.Load(path)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if err := writeListResponse(streams.Out, cfg); err != nil {
		return err
	}

	return nil
}

func writeListResponse(out io.Writer, resp config.Config) error {
	if len(resp.Services) == 0 {
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

	names := make([]string, 0, len(resp.Services))
	for name := range resp.Services {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		service := resp.Services[name]
		if _, err := fmt.Fprintf(
			writer,
			"%s\t%s\t%s\t%d\t%t\n",
			name,
			service.FQDN,
			service.IP,
			service.Port,
			service.Enabled,
		); err != nil {
			return err
		}
	}

	return writer.Flush()
}
