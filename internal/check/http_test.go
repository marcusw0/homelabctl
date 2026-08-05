package check

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckHTTP(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name        string
		status      int
		wantHealthy bool
	}{
		{"success", http.StatusOK, true},
		{"last healthy status", 299, true},
		{"redirect boundary", 300, false},
		{"server error", http.StatusInternalServerError, false},
	}

	s := HTTP{
		Timeout: 3 * time.Second,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.status)
				},
			))

			defer server.Close()
			got, err := s.Check(ctx, server.URL)
			if err != nil {
				t.Error(err)
			}

			if got.StatusCode != tt.status {
				t.Errorf("Got: %d\nExpected: %d\n", got.StatusCode, tt.status)
			}

			if got.Healthy != tt.wantHealthy {
				t.Errorf("Got: %v\nExpected: %v\n", got.Healthy, tt.wantHealthy)
			}
		})
	}
}
