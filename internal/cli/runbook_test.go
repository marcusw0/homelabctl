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

func TestRunbookCmdRun(t *testing.T) {
	dir := t.TempDir()
	runbookDir := filepath.Join(dir, "runbooks")

	if err := os.Mkdir(runbookDir, 0o700); err != nil {
		t.Fatal(err)
	}

	runbookPath := filepath.Join(runbookDir, "myserver.md")

	if err := os.WriteFile(
		runbookPath,
		[]byte("# MyServer Runbook\n\nCheck the logs."),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	cmd := RunbookCmd{
		RunbookPath: runbookPath,
		ServerName:  "myserver",
		Style:       "notty",
		Width:       80,
	}

	var output bytes.Buffer

	err := cmd.Run(
		context.Background(),
		IOStreams{
			In:     strings.NewReader(""),
			Out:    &output,
			ErrOut: io.Discard,
		},
	)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	got := output.String()

	if !strings.Contains(got, "MyServer Runbook") {
		t.Errorf("output does not contain heading:\n%s", got)
	}

	if !strings.Contains(got, "Check the logs") {
		t.Errorf("output does not contain runbook text:\n%s", got)
	}
}
