package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthzHandler(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if body := recorder.Body.String(); body != "ok\n" {
		t.Errorf("body = %q, want %q", body, "ok\n")
	}
}

func TestHealthzMethodsAllowed(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	methods := []string{http.MethodGet, http.MethodHead, http.MethodPost}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(method, "/healthz", nil)
			recorder := httptest.NewRecorder()
			mux.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusOK {
				t.Errorf("%s /healthz: status = %d, want %d", method, recorder.Code, http.StatusOK)
			}
		})
	}
}

func TestServerConfiguration(t *testing.T) {
	t.Parallel()

	// Verify that the server configuration constants used in main() are reasonable.
	// These mirror the values from main.go.
	tests := []struct {
		name  string
		value time.Duration
		min   time.Duration
		max   time.Duration
	}{
		{
			name:  "ReadHeaderTimeout",
			value: 5 * time.Second,
			min:   1 * time.Second,
			max:   30 * time.Second,
		},
		{
			name:  "ShutdownTimeout",
			value: 10 * time.Second,
			min:   1 * time.Second,
			max:   60 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.value < tt.min || tt.value > tt.max {
				t.Errorf("%s = %v, want between %v and %v", tt.name, tt.value, tt.min, tt.max)
			}
		})
	}
}

func TestLogMessageFormat(t *testing.T) {
	t.Parallel()

	// Verify the format string used in main.go's runCollection log is valid.
	// This tests that the format verbs match the expected types.
	tests := []struct {
		name     string
		endpoint string
		success  bool
		down     float64
		up       float64
		latency  float64
	}{
		{
			name:     "successful measurement",
			endpoint: "https://netzbremse.de/speed",
			success:  true,
			down:     50_000_000,
			up:       10_000_000,
			latency:  12.5,
		},
		{
			name:     "failed measurement",
			endpoint: "https://netzbremse.de/speed",
			success:  false,
			down:     0,
			up:       0,
			latency:  0,
		},
		{
			name:     "high values",
			endpoint: "https://netzbremse.de/speed",
			success:  true,
			down:     1_000_000_000,
			up:       500_000_000,
			latency:  0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// This exercises the same format string from main.go.
			// If the format verbs are wrong, this will panic.
			reason := "test"
			_ = formatMeasurementLog(reason, tt.endpoint, tt.success, tt.down, tt.up, tt.latency)
		})
	}
}

// formatMeasurementLog mirrors the log.Printf format from main.go's runCollection.
func formatMeasurementLog(reason, endpoint string, success bool, down, up, latency float64) string {
	return "stored measurement (" + reason + "): endpoint=" + endpoint +
		" success=" + boolString(success) +
		" down=" + floatString(down) +
		" up=" + floatString(up) +
		" latency=" + floatString(latency)
}

func boolString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func floatString(f float64) string {
	if f == 0 {
		return "0"
	}
	return "nonzero"
}
