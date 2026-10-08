package cli

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"text/tabwriter"

	"github.com/marcusw0/homelabctl/internal/config"
	"github.com/marcusw0/homelabctl/internal/runner"
)

func runAllCheck(ctx context.Context, path string, streams IOStreams) error {
	cfg, err := config.Load(path)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	enabled := 0
	for _, service := range cfg.Services {
		if service.Enabled {
			enabled++
		}
	}
	if enabled == 0 {
		return errors.New("no enabled services configured")
	}

	const maxConcurrent = 5

	jobs := make([]runner.Job, 0, enabled)

	for name, service := range cfg.Services {
		if !service.Enabled {
			continue
		}
		s := serviceFromConfig(service)

		jobs = append(jobs, runner.Job{
			Name:    name,
			Checker: &s,
		})
	}

	serviceRunner := runner.Runner{
		MaxConcurrent: maxConcurrent,
	}

	results := make([]runner.Result, 0, len(jobs))
	var errs []error

	for result := range serviceRunner.Run(ctx, jobs) {
		results = append(results, result)
		if result.Err != nil {
			errs = append(errs, fmt.Errorf("service %q: %w", result.Name, result.Err))
		}
	}

	errs = append(errs, ctx.Err(), writeAllResults(streams.ErrOut, streams.Out, results))
	return errors.Join(errs...)
}

func writeAllResults(errOut io.Writer, out io.Writer, results []runner.Result) error {
	slices.SortFunc(results, func(a, b runner.Result) int {
		return cmp.Compare(a.Name, b.Name)
	})

	writer := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(
		writer,
		"NAME\tHTTP\tDNS\tTCP\tTLS",
	); err != nil {
		return err
	}

	for _, v := range results {
		if v.Err != nil {
			fmt.Fprintf(errOut, "[ERROR]%s returned: %v\n", v.Name, v.Err)
		}
		if _, err := fmt.Fprintf(
			writer,
			"%s\t%s\t%s\t%s\t%s\n",
			v.Name,
			v.Checks.HTTP.Status,
			v.Checks.DNS.Status,
			v.Checks.TCP.Status,
			v.Checks.TLS.Status,
		); err != nil {
			return err
		}
	}
	return writer.Flush()
}
