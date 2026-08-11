package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/marcusw0/homelabctl/internal/check"
)

type HTTPCheckCmd struct {
	Target         string
	Timeout        time.Duration
	ExpectedStatus int
	FollowRedirect bool
	Verbose        bool
}

func parseHTTPCheck(
	errOut io.Writer,
	args []string,
	opts GlobalOption,
) (Command, error) {
	cmd := &HTTPCheckCmd{}

	flags := flag.NewFlagSet("check http", flag.ContinueOnError)
	flags.SetOutput(errOut)
	addGlobalFlags(flags, &opts)

	flags.DurationVar(
		&cmd.Timeout,
		"timeout",
		5*time.Second,
		"HTTP Request timeout",
	)

	flags.IntVar(
		&cmd.ExpectedStatus,
		"expect-status",
		http.StatusOK,
		"Expected HTTP Status Code",
	)

	flags.BoolVar(
		&cmd.FollowRedirect,
		"follow-redirects",
		true,
		"Should redirects be followed",
	)

	if err := flags.Parse(args); err != nil {
		return nil, err
	}
	cmd.Verbose = opts.Verbose

	if flags.NArg() == 1 {
		cmd.Target = flags.Arg(0)
	}

	if flags.NArg() > 1 {
		return nil, errors.New("check http accepts only one argument")
	}

	return cmd, nil
}

func (c *HTTPCheckCmd) Validate() error {
	if c.Target == "" {
		return errors.New("Need to specify destination to check")
	}

	if c.Timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}

	if c.ExpectedStatus < 100 || c.ExpectedStatus > 599 {
		return errors.New("expected-status must be between 100 and 599")
	}
	if !strings.HasPrefix(c.Target, "https://") &&
		!strings.HasPrefix(c.Target, "http://") {
		c.Target = "https://" + c.Target
	}

	parsedURL, err := url.ParseRequestURI(c.Target)
	if err != nil || parsedURL.Host == "" {
		return fmt.Errorf("invalid HTTP target %q", c.Target)
	}

	return nil
}

func (c *HTTPCheckCmd) Run(ctx context.Context, streams IOStreams) error {
	service := check.HTTP{
		Timeout:        c.Timeout,
		ExpectedStatus: c.ExpectedStatus,
		FollowRedirect: c.FollowRedirect,
	}

	resp, respErr := service.Check(ctx, c.Target)
	writeErr := writeHTTPResponse(streams.Out, resp, c.Verbose)
	if writeErr != nil {
		return errors.Join(respErr, writeErr)
	}
	return respErr
}

func writeHTTPResponse(
	out io.Writer,
	resp check.HTTPResults,
	verbose bool,
) error {
	if !verbose {
		if resp.StatusCode == 0 {
			_, err := fmt.Fprintf(
				out,
				"Server: %s\nHealthy: %t\n",
				resp.Target,
				resp.Healthy,
			)
			return err
		}

		_, err := fmt.Fprintf(
			out,
			"Server: %s\nStatus: %d\nHealthy: %t\n",
			resp.Target,
			resp.StatusCode,
			resp.Healthy,
		)
		return err
	}

	if resp.StatusCode == 0 {
		_, err := fmt.Fprintf(
			out,
			"Server: %s\nHealthy: %t\nChecked At: %s\n",
			resp.Target,
			resp.Healthy,
			formatTimestamp(resp.CheckedAt),
		)
		return err
	}

	_, err := fmt.Fprintf(
		out,
		"Server: %s\n"+
			"Status: %d\n"+
			"Latency: %s\n"+
			"Healthy: %t\n"+
			"Checked At: %s\n",
		resp.Target,
		resp.StatusCode,
		formatDuration(resp.Latency),
		resp.Healthy,
		formatTimestamp(resp.CheckedAt),
	)
	if err != nil {
		return err
	}
	return nil
}
