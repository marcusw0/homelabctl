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
}

func parseHTTPCheck(errOut io.Writer, args []string, opts GlobalOption) (Command, error) {
	cmd := &HTTPCheckCmd{}

	flags := flag.NewFlagSet("check http", flag.ContinueOnError)
	flags.SetOutput(errOut)

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
	if respErr != nil {
		_, writeErr := fmt.Fprintf(
			streams.Out,
			"FAILED\nHost: %s\nHealthy: %t\nChecked At: %v\n",
			resp.Target,
			resp.Healthy,
			resp.CheckedAt.Local(),
		)
		return errors.Join(respErr, writeErr)
	}

	if err := writeHTTPResponse(streams.Out, resp); err != nil {
		return err
	}
	return nil
}

func writeHTTPResponse(out io.Writer, resp check.HTTPResults) error {
	_, err := fmt.Fprintf(
		out,
		"Server: %s\nStatus: %d\nLatency: %dms\nHealthy: %t\nChecked At: %v\n",
		resp.Target,
		resp.StatusCode,
		resp.Latency.Milliseconds(),
		resp.Healthy,
		resp.CheckedAt.Local(),
	)
	if err != nil {
		return err
	}
	return nil
}
