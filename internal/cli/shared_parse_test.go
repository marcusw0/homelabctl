package cli

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"strings"
	"testing"
	"time"
)

func TestSharedParsersRequestData(t *testing.T) {
	commands := []struct {
		family string
		name   string
		kind   CommandKind
	}{
		{"check", "dns", CmdDNSCheck},
		{"check", "tcp", CmdTCPCheck},
		{"check", "tls", CmdTLSCheck},
		{"check", "service", CmdServiceCheck},
		{"config", "add", CmdAddConfig},
		{"config", "edit", CmdEditConfig},
	}
	for _, command := range commands {
		positions := []string{"omitted"}
		if command.kind == CmdServiceCheck || command.kind == CmdAddConfig {
			positions = append(positions, "root", "family", "subcommand")
		}
		for _, position := range positions {
			t.Run(command.family+"/"+command.name+"/"+position, func(t *testing.T) {
				args := []string{"--config", "base.toml"}
				options := []string{"--config", "override.toml", "--timeout", "2s", "--verbose"}
				if position == "root" {
					args = append(args, options...)
				}
				args = append(args, command.family)
				if position == "family" {
					args = append(args, options...)
				}
				args = append(args, command.name)
				if position == "subcommand" {
					args = append(args, options...)
				}
				args = append(args, "web")

				request, err := Parse(args, io.Discard)
				if err != nil {
					t.Fatal(err)
				}
				if request.Kind != command.kind {
					t.Fatalf("kind = %d, want %d", request.Kind, command.kind)
				}
				wantPath := "base.toml"
				if position != "omitted" {
					wantPath = "override.toml"
				}
				if command.family == "config" {
					if request.Modify.ServiceName != "web" || request.Modify.ConfigPath != wantPath {
						t.Fatalf("unexpected modification request: %+v", request.Modify)
					}
					return
				}
				if request.Check.Target != "web" || request.Check.Verbose != (position != "omitted") {
					t.Fatalf("unexpected check request: %+v", request.Check)
				}
				if position == "omitted" {
					if request.Check.Timeout != nil {
						t.Fatal("omitted timeout must remain nil")
					}
				} else if request.Check.Timeout == nil || *request.Check.Timeout != 2*time.Second {
					t.Fatal("timeout override was not preserved")
				}
				if command.kind != CmdServiceCheck {
					wantPath = ""
				}
				if request.ConfigPath != wantPath {
					t.Fatalf("config path = %q, want %q", request.ConfigPath, wantPath)
				}
			})
		}
	}
}

func TestSharedParsersRejectArgumentsAndShowHelp(t *testing.T) {
	for _, command := range []string{"check service", "config add"} {
		for _, arguments := range [][]string{nil, {"web", "extra"}, {"--unknown"}, {"--help"}} {
			t.Run(command+"/"+strings.Join(arguments, " "), func(t *testing.T) {
				args := append(strings.Fields(command), arguments...)
				var output bytes.Buffer
				_, err := Parse(args, &output)
				if err == nil {
					t.Fatal("expected an error")
				}
				if len(arguments) == 1 && arguments[0] == "--help" {
					if !errors.Is(err, flag.ErrHelp) || !strings.Contains(output.String(), command) {
						t.Fatalf("help error = %v, output = %q", err, output.String())
					}
				}
			})
		}
	}
}
