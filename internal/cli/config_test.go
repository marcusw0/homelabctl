package cli

import (
	"io"
	"testing"
)

func TestParseConfigAdd(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"missing server name", []string{"add"}, true},
		{"server name", []string{"add", "gitlab"}, false},
		{"too many arguments", []string{"add", "gitlab", "extra"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseConfig(
				io.Discard,
				tt.args,
				GlobalOption{ConfigPath: "config.toml"},
			)

			gotErr := err != nil
			if gotErr != tt.wantErr {
				t.Errorf("Got error: %v\nExpected error: %t\n", err, tt.wantErr)
			}
		})
	}
}

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

			cmd, ok := parsed.(*ConfigInitCmd)
			if !ok {
				t.Fatalf("unexpected command type %T", parsed)
			}

			want := tt.name + ".toml"
			if cmd.ConfigPath != want {
				t.Fatalf("config path = %q, want %q", cmd.ConfigPath, want)
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

			cmd, ok := parsed.(*HTTPCheckCmd)
			if !ok {
				t.Fatalf("unexpected command type %T", parsed)
			}
			if !cmd.Verbose {
				t.Fatal("verbose flag was not added to HTTP command")
			}
		})
	}
}
