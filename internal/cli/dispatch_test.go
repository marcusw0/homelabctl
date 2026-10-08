package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDispatchConfigInitAndList(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.toml")
	var output bytes.Buffer
	streams := IOStreams{Out: &output, ErrOut: io.Discard}
	ctx := context.Background()

	if err := Dispatch(ctx, Request{Kind: CmdInitConfig, ConfigPath: path}, streams); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "config initialized at "+path) {
		t.Fatalf("unexpected initialization output: %q", output.String())
	}
	output.Reset()
	if err := Dispatch(ctx, Request{Kind: CmdListConfig, ConfigPath: path}, streams); err != nil {
		t.Fatal(err)
	}
	if output.String() != "No services configured.\n" {
		t.Fatalf("unexpected list output: %q", output.String())
	}
	if err := Dispatch(ctx, Request{Kind: CmdInitConfig, ConfigPath: path}, streams); err == nil {
		t.Fatal("initializing an existing file should fail")
	}
	content, err := os.ReadFile(path)
	if err != nil || len(content) != 0 {
		t.Fatalf("existing configuration changed: content=%q, error=%v", content, err)
	}
	if err := Dispatch(ctx, Request{Kind: CmdAllCheck, ConfigPath: path}, streams); err == nil || !strings.Contains(err.Error(), "no enabled services") {
		t.Fatalf("empty inventory error = %v", err)
	}
}

func TestDispatchPreparationErrors(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.toml")
	cases := []struct {
		name    string
		request Request
		message string
	}{
		{"edit", Request{Kind: CmdEditConfig, Modify: ModifyConfigRequest{ConfigPath: missing, ServiceName: "web"}}, "load config"},
		{"dashboard", Request{Kind: CmdDashboard, ConfigPath: missing}, "load config"},
		{"unknown", Request{}, "unsupported command kind"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := Dispatch(context.Background(), tt.request, IOStreams{Out: io.Discard, ErrOut: io.Discard})
			if err == nil || !strings.Contains(err.Error(), tt.message) {
				t.Fatalf("error = %v, want %q", err, tt.message)
			}
		})
	}
}
