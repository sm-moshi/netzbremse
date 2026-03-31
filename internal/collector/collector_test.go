package collector

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/sm-moshi/netzbremse/internal/config"
)

func TestRunEmptyCommand(t *testing.T) {
	t.Parallel()

	cfg := config.Measurement{
		Command:  "",
		Timeout:  10 * time.Second,
		Endpoint: "https://example.com",
	}

	_, err := Run(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error for empty command")
	}
}

func TestRunSuccessfulCommand(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("test requires unix echo command")
	}

	// Use echo to output valid speedtest JSON.
	payload := `{"sessionID":"test-123","endpoint":"https://test.example.com","success":true,"timestamp":"2026-03-15T10:00:00Z","result":{"download":50000000,"upload":10000000,"latency":12.5,"jitter":3.2,"downLoadedLatency":15,"downLoadedJitter":4,"upLoadedLatency":18,"upLoadedJitter":5}}`

	cfg := config.Measurement{
		Command:  "echo " + payload,
		Timeout:  10 * time.Second,
		Endpoint: "https://fallback.example.com",
	}

	m, err := Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.SessionID != "test-123" {
		t.Errorf("SessionID = %q, want %q", m.SessionID, "test-123")
	}
	if m.Endpoint != "https://test.example.com" {
		t.Errorf("Endpoint = %q, want %q", m.Endpoint, "https://test.example.com")
	}
	if !m.Success {
		t.Error("Success = false, want true")
	}
	if m.DownloadBPS == nil || *m.DownloadBPS != 50_000_000 {
		t.Errorf("DownloadBPS = %v, want 50000000", m.DownloadBPS)
	}
	if m.UploadBPS == nil || *m.UploadBPS != 10_000_000 {
		t.Errorf("UploadBPS = %v, want 10000000", m.UploadBPS)
	}
	if m.LatencyMS == nil || *m.LatencyMS != 12.5 {
		t.Errorf("LatencyMS = %v, want 12.5", m.LatencyMS)
	}
}

func TestRunUseFallbackEndpoint(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("test requires unix echo command")
	}

	// Payload with no endpoint — should use fallback.
	payload := `{"success":true,"result":{"download":1000,"upload":500,"latency":10,"jitter":1}}`

	cfg := config.Measurement{
		Command:  "echo " + payload,
		Timeout:  10 * time.Second,
		Endpoint: "https://fallback.example.com",
	}

	m, err := Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Endpoint != "https://fallback.example.com" {
		t.Errorf("Endpoint = %q, want %q", m.Endpoint, "https://fallback.example.com")
	}
}

func TestRunCommandFailure(t *testing.T) {
	t.Parallel()

	cfg := config.Measurement{
		Command:  "false",
		Timeout:  10 * time.Second,
		Endpoint: "https://example.com",
	}

	_, err := Run(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error for failing command")
	}
}

func TestRunCommandTimeout(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("test requires unix sleep command")
	}

	cfg := config.Measurement{
		Command:  "sleep 60",
		Timeout:  50 * time.Millisecond,
		Endpoint: "https://example.com",
	}

	start := time.Now()
	_, err := Run(context.Background(), cfg)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error for timed-out command")
	}

	// Should return well before the sleep duration.
	if elapsed > 5*time.Second {
		t.Errorf("timeout took too long: %v", elapsed)
	}
}

func TestRunCancelledContext(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("test requires unix sleep command")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Already cancelled.

	cfg := config.Measurement{
		Command:  "sleep 60",
		Timeout:  30 * time.Second,
		Endpoint: "https://example.com",
	}

	_, err := Run(ctx, cfg)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestRunInvalidJSON(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("test requires unix echo command")
	}

	cfg := config.Measurement{
		Command:  "echo not-json",
		Timeout:  10 * time.Second,
		Endpoint: "https://example.com",
	}

	_, err := Run(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error for invalid JSON output")
	}
}

func TestRunMultiWordCommand(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("test requires unix printf command")
	}

	// printf with format + arg produces valid JSON.
	cfg := config.Measurement{
		Command:  `printf {"success":true,"result":{"download":1,"upload":1,"latency":1,"jitter":1}}`,
		Timeout:  10 * time.Second,
		Endpoint: "https://example.com",
	}

	m, err := Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !m.Success {
		t.Error("Success = false, want true")
	}
}

func TestRunNonExistentCommand(t *testing.T) {
	t.Parallel()

	cfg := config.Measurement{
		Command:  "/nonexistent/binary/path",
		Timeout:  10 * time.Second,
		Endpoint: "https://example.com",
	}

	_, err := Run(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error for non-existent command")
	}
}
