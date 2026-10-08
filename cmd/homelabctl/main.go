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

	request, err := cli.Parse(os.Args[1:], streams.ErrOut)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintf(streams.ErrOut, "ERROR: %v\n", err)
		os.Exit(2)
	}

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	if err := cli.Dispatch(ctx, request, streams); err != nil {
		fmt.Fprintf(streams.ErrOut, "ERROR: %v\n", err)
		os.Exit(1)
	}
}
