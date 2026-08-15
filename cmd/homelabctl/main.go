package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/marcusw0/homelabctl/internal/cli"
)

func main() {

	streams := cli.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	if len(os.Args) <= 1 {
		fmt.Fprintln(
			streams.ErrOut,
			"Usage: homelabctl check <http|tcp|tls|dns> <target>",
		)
		os.Exit(2)
	}

	parsed, err := cli.Parse(os.Args[1:], streams.ErrOut)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintf(streams.ErrOut, "ERROR: %v\n", err)
		os.Exit(2)
	}

	if err := parsed.Validate(); err != nil {
		fmt.Fprintf(streams.ErrOut, "ERROR: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	if err := parsed.Run(ctx, streams); err != nil {
		fmt.Fprintf(streams.ErrOut, "ERROR: %v\n", err)
		os.Exit(1)
	}

}
