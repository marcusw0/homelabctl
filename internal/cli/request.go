package cli

import (
	"flag"
	"fmt"
	"io"
	"time"
)

type CommandKind uint8

const (
	CmdHTTPCheck CommandKind = iota + 1
	CmdServiceCheck
	CmdTLSCheck
	CmdDNSCheck
	CmdTCPCheck
	CmdAllCheck
	CmdAddConfig
	CmdEditConfig
	CmdInitConfig
	CmdListConfig
	CmdDashboard
	CmdRunbook
)

type Request struct {
	HTTP    HTTPCheckRequest
	Runbook RunbookRequest
	Check   CheckRequest
	Modify  ModifyConfigRequest

	ConfigPath string
	Kind       CommandKind
}

type HTTPCheckRequest struct {
	Target         string
	Timeout        *time.Duration
	FollowRedirect bool
	Verbose        bool
	ExpectedStatus int
}

type RunbookRequest struct {
	ConfigPath  string
	ServiceName string
	Style       string
	Width       int
}

type CheckRequest struct {
	Target  string
	Timeout *time.Duration
	Verbose bool
}

type ModifyConfigRequest struct {
	ConfigPath  string
	ServiceName string
}

type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}

const defaultTimeout = 5 * time.Second

type GlobalOption struct {
	ConfigPath string
	Timeout    *time.Duration
	Verbose    bool
}

func addGlobalFlags(flags *flag.FlagSet, opts *GlobalOption) {
	flags.StringVar(
		&opts.ConfigPath,
		"config",
		opts.ConfigPath,
		"path to config file",
	)

	flags.BoolVar(
		&opts.Verbose,
		"v",
		opts.Verbose,
		"verbose output",
	)

	flags.BoolVar(
		&opts.Verbose,
		"verbose",
		opts.Verbose,
		"verbose output",
	)

	flags.Func(
		"timeout",
		"override configured timeout",
		func(value string) error {
			timeout, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid timeout: %w", err)
			}

			opts.Timeout = &timeout
			return nil
		},
	)
}
