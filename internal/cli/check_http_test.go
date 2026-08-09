package cli

import (
	"net/http"
	"testing"
	"time"
)

func TestHTTPCheckCmdValidate(t *testing.T) {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := HTTPCheckCmd{
				Target:         tt.target,
				Timeout:        time.Second,
				ExpectedStatus: http.StatusOK,
			}

			err := cmd.Validate()
			gotErr := err != nil
			if gotErr != tt.wantErr {
				t.Errorf("Got error: %v\nExpected error: %t\n", err, tt.wantErr)
			}

			if cmd.Target != tt.wantTarget {
				t.Errorf("Got: %s\nExpected: %s\n", cmd.Target, tt.wantTarget)
			}
		})
	}
}
