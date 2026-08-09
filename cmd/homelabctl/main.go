package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

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
	}

	parsed, err := cli.Parse(os.Args[1:], streams.ErrOut)
	if err != nil {
		fmt.Fprintf(streams.ErrOut, "ERORR %v", err)
		os.Exit(2)
	}

	if err := parsed.Validate(); err != nil {
		fmt.Fprintf(streams.ErrOut, "ERORR %v", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
	)
	defer cancel()

	if err := parsed.Run(ctx, streams); err != nil {
		fmt.Fprintf(streams.ErrOut, "ERORR %v", err)
		os.Exit(1)
	}

	os.Exit(0)
}
