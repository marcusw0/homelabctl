package cli

import (
	"io"
	"testing"
)

func TestGlobalConfigFlagPositions(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"root", []string{"--config", "root.toml", "config", "init"}},
		{"command", []string{"config", "--config", "command.toml", "init"}},
		{"subcommand", []string{"config", "init", "--config", "subcommand.toml"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := Parse(tt.args, io.Discard)
			if err != nil {
				t.Fatalf("parse command: %v", err)
			}

			if parsed.Kind != CmdInitConfig {
				t.Fatalf("unexpected command kind %d", parsed.Kind)
			}

			want := tt.name + ".toml"
			if parsed.ConfigPath != want {
				t.Fatalf("config path = %q, want %q", parsed.ConfigPath, want)
			}
		})
	}
}

func TestGlobalVerboseFlagPositions(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"root", []string{"-v", "check", "http", "example.com"}},
		{"subcommand", []string{"check", "http", "--verbose", "example.com"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := Parse(tt.args, io.Discard)
			if err != nil {
				t.Fatalf("parse command: %v", err)
			}

			if parsed.Kind != CmdHTTPCheck {
				t.Fatalf("unexpected command kind %d", parsed.Kind)
			}
			if !parsed.HTTP.Verbose {
				t.Fatal("verbose flag was not added to HTTP command")
			}
		})
	}
}
