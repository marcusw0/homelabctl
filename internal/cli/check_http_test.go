package cli

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestHTTPCheckRequest(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		wantTarget string
		wantErr    bool
	}{
		{"adds https", "example.com", "https://example.com", false},
		{"keeps http", "http://example.com", "http://example.com", false},
		{"keeps https", "https://example.com", "https://example.com", false},
		{"malformed URL", "https://", "https://", true},
	}

	original := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = original })
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			http.DefaultTransport = testTransport(func(r *http.Request) (*http.Response, error) {
				called = true
				if r.URL.String() != tt.wantTarget {
					t.Errorf("request URL = %q, want %q", r.URL.String(), tt.wantTarget)
				}
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
			})
			timeout := time.Second
			request := HTTPCheckRequest{
				Target:         tt.target,
				Timeout:        &timeout,
				ExpectedStatus: http.StatusOK,
			}

			err := Dispatch(context.Background(), Request{Kind: CmdHTTPCheck, HTTP: request}, IOStreams{Out: io.Discard, ErrOut: io.Discard})
			gotErr := err != nil
			if gotErr != tt.wantErr {
				t.Errorf("Got error: %v\nExpected error: %t\n", err, tt.wantErr)
			}

			if called == tt.wantErr {
				t.Errorf("HTTP called = %t, want %t", called, !tt.wantErr)
			}
		})
	}
}
