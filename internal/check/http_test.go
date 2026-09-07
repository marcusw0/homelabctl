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
		expected    int
		wantHealthy bool
	}{
		{"success", http.StatusOK, http.StatusOK, true},
		{"expected no content", http.StatusNoContent, http.StatusNoContent, true},
		{"redirect boundary", 300, 300, true},
		{"expected client error", http.StatusNotFound, http.StatusNotFound, true},
		{"expected server error", http.StatusInternalServerError, http.StatusInternalServerError, true},
		{"unexpected client error", http.StatusNotFound, http.StatusOK, false},
		{"unexpected server error", http.StatusInternalServerError, http.StatusOK, false},
		{"unexpected success", http.StatusOK, http.StatusNoContent, false},
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
			s.ExpectedStatus = tt.expected
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
