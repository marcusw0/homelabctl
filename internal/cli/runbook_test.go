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

func TestRunbookDispatch(t *testing.T) {
	dir := t.TempDir()
	runbookDir := filepath.Join(dir, "runbooks")

	if err := os.Mkdir(runbookDir, 0o700); err != nil {
		t.Fatal(err)
	}

	runbookPath := filepath.Join(runbookDir, "myservice.md")

	if err := os.WriteFile(
		runbookPath,
		[]byte("# MyService Runbook\n\nCheck the logs."),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(configPath, []byte(`[services.myservice]
fqdn = "localhost"
ip = "127.0.0.1"
port = 443
runbook = "runbooks/myservice.md"
`), 0o600); err != nil {
		t.Fatal(err)
	}

	request := RunbookRequest{
		ConfigPath:  configPath,
		ServiceName: "myservice",
		Style:       "notty",
		Width:       80,
	}

	var output bytes.Buffer

	err := Dispatch(
		context.Background(),
		Request{Kind: CmdRunbook, Runbook: request},
		IOStreams{
			In:     strings.NewReader("q"),
			Out:    &output,
			ErrOut: io.Discard,
		},
	)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
}
