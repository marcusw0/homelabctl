package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/marcusw0/homelabctl/internal/cli"
)

func main() {
	if len(os.Args) <= 1 {
		log.Fatalln("Usage: homelabctl check <http|tcp|tls|dns> <target>")
	}

	parsed, err := cli.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("Fatal Error: %q\n", err)
	}

	if err := parsed.Validate(); err != nil {
		log.Fatalf("Fatal Error: %q\n", err)
	}

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
	)
	defer cancel()

	if err := parsed.Run(ctx); err != nil {
		log.Fatalf("Fatal Error: %q\n", err)
	}

	log.Println("done")
	os.Exit(0)
}
